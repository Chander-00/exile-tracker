window.renderPassiveTree = function (containerId, allocatedNodeIds) {
  var container = document.getElementById(containerId);
  if (!container) return;

  var allocSet = {};
  for (var i = 0; i < allocatedNodeIds.length; i++) {
    allocSet[allocatedNodeIds[i]] = true;
  }

  fetch("/static/tree.json")
    .then(function (r) { return r.json(); })
    .then(function (data) {
      var groups = data.groups;
      var nodes = data.nodes;
      var skillsPerOrbit = data.constants.skillsPerOrbit;
      var orbitRadii = data.constants.orbitRadii;

      // Compute all node positions
      var nodePos = {};
      var nodeKeys = Object.keys(nodes);
      for (var i = 0; i < nodeKeys.length; i++) {
        var id = nodeKeys[i];
        var node = nodes[id];
        var group = groups[node.g];
        if (!group) continue;
        if (node.o === 0) {
          nodePos[id] = { x: group.x, y: group.y };
        } else {
          var radius = orbitRadii[node.o] || 0;
          var total = skillsPerOrbit[node.o] || 1;
          var angle = (2 * Math.PI * node.oi) / total - Math.PI / 2;
          nodePos[id] = {
            x: group.x + radius * Math.cos(angle),
            y: group.y + radius * Math.sin(angle)
          };
        }
      }

      // Compute bounding box of allocated nodes for initial zoom
      var allocMinX = Infinity, allocMinY = Infinity;
      var allocMaxX = -Infinity, allocMaxY = -Infinity;
      var hasAlloc = false;
      for (var i = 0; i < allocatedNodeIds.length; i++) {
        var pos = nodePos[allocatedNodeIds[i]];
        if (!pos) continue;
        hasAlloc = true;
        if (pos.x < allocMinX) allocMinX = pos.x;
        if (pos.y < allocMinY) allocMinY = pos.y;
        if (pos.x > allocMaxX) allocMaxX = pos.x;
        if (pos.y > allocMaxY) allocMaxY = pos.y;
      }

      // Viewbox state for zoom/pan
      var view = {};
      if (hasAlloc) {
        var pad = 800;
        view.x = allocMinX - pad;
        view.y = allocMinY - pad;
        view.w = (allocMaxX - allocMinX) + pad * 2;
        view.h = (allocMaxY - allocMinY) + pad * 2;
      } else {
        view.x = data.min_x - 500;
        view.y = data.min_y - 500;
        view.w = (data.max_x - data.min_x) + 1000;
        view.h = (data.max_y - data.min_y) + 1000;
      }

      // Scale node sizes relative to view (so they're visible at any zoom)
      var NODE_SCALE = 1; // will be updated
      function getNodeRadius(type) {
        var base = { notable: 40, keystone: 60, mastery: 30, jewel: 40, ascStart: 40 };
        return (base[type] || 20) * NODE_SCALE;
      }
      function getStrokeWidth() { return 12 * NODE_SCALE; }
      function getEdgeWidth(highlight) { return (highlight ? 16 : 8) * NODE_SCALE; }

      var ns = "http://www.w3.org/2000/svg";
      var svg = document.createElementNS(ns, "svg");
      svg.setAttribute("width", "100%");
      svg.setAttribute("height", "100%");
      svg.style.background = "#0a0a0f";
      svg.style.borderRadius = "6px";
      svg.style.cursor = "grab";

      function updateViewBox() {
        svg.setAttribute("viewBox", view.x + " " + view.y + " " + view.w + " " + view.h);
        NODE_SCALE = view.w / 5000; // scale nodes relative to visible area
        if (NODE_SCALE < 0.3) NODE_SCALE = 0.3;
        if (NODE_SCALE > 3) NODE_SCALE = 3;
      }
      updateViewBox();

      // Draw edges
      var edgeGroup = document.createElementNS(ns, "g");
      edgeGroup.setAttribute("class", "edges");
      for (var i = 0; i < nodeKeys.length; i++) {
        var id = nodeKeys[i];
        var node = nodes[id];
        var from = nodePos[id];
        if (!from || !node.out) continue;
        for (var j = 0; j < node.out.length; j++) {
          var to = nodePos[node.out[j]];
          if (!to) continue;
          var bothAlloc = allocSet[id] && allocSet[node.out[j]];
          var line = document.createElementNS(ns, "line");
          line.setAttribute("x1", from.x);
          line.setAttribute("y1", from.y);
          line.setAttribute("x2", to.x);
          line.setAttribute("y2", to.y);
          line.setAttribute("stroke", bothAlloc ? "#6b5a00" : "#1e1e1e");
          line.setAttribute("stroke-width", bothAlloc ? 16 : 8);
          edgeGroup.appendChild(line);
        }
      }
      svg.appendChild(edgeGroup);

      // Draw nodes
      var nodeGroup = document.createElementNS(ns, "g");
      nodeGroup.setAttribute("class", "nodes");
      for (var i = 0; i < nodeKeys.length; i++) {
        var id = nodeKeys[i];
        var node = nodes[id];
        var pos = nodePos[id];
        if (!pos) continue;
        var isAlloc = !!allocSet[id];
        var baseR = { notable: 40, keystone: 60, mastery: 30, jewel: 40, ascStart: 40 };
        var r = baseR[node.t] || 20;
        var circle = document.createElementNS(ns, "circle");
        circle.setAttribute("cx", pos.x);
        circle.setAttribute("cy", pos.y);
        circle.setAttribute("r", r);
        circle.setAttribute("fill", isAlloc ? "#AF8700" : "#2a2a2a");
        if (isAlloc) {
          circle.setAttribute("stroke", "#d4a017");
          circle.setAttribute("stroke-width", "12");
        }
        circle.setAttribute("data-node-id", id);
        // Larger invisible hit area for hover
        var hitArea = document.createElementNS(ns, "circle");
        hitArea.setAttribute("cx", pos.x);
        hitArea.setAttribute("cy", pos.y);
        hitArea.setAttribute("r", r * 3);
        hitArea.setAttribute("fill", "transparent");
        hitArea.setAttribute("data-node-id", id);
        hitArea.style.cursor = "pointer";
        nodeGroup.appendChild(circle);
        if (node.n) nodeGroup.appendChild(hitArea);
      }
      svg.appendChild(nodeGroup);

      container.style.position = "relative";
      container.style.overflow = "hidden";
      container.appendChild(svg);

      // Tooltip
      var tooltip = document.createElement("div");
      tooltip.style.cssText =
        "position:absolute;display:none;background:#1a1a2e;border:1px solid #AF8700;" +
        "color:#ccc;padding:8px 12px;border-radius:4px;font-size:13px;pointer-events:none;" +
        "z-index:10;max-width:320px;white-space:pre-wrap;line-height:1.5;box-shadow:0 4px 12px rgba(0,0,0,0.5);";
      container.appendChild(tooltip);

      // Hover events (using event delegation)
      var currentHover = null;
      svg.addEventListener("mouseover", function (e) {
        var t = e.target;
        if (t.tagName !== "circle") return;
        var nid = t.getAttribute("data-node-id");
        if (!nid || nid === currentHover) return;
        currentHover = nid;
        var node = nodes[nid];
        if (!node || !node.n) { tooltip.style.display = "none"; return; }
        var text = node.n;
        if (node.s && node.s.length > 0) {
          text += "\n\n" + node.s.join("\n");
        }
        tooltip.textContent = text;
        tooltip.style.display = "block";
      });

      svg.addEventListener("mousemove", function (e) {
        if (tooltip.style.display === "none") return;
        var rect = container.getBoundingClientRect();
        var x = e.clientX - rect.left + 15;
        var y = e.clientY - rect.top - 10;
        // Keep tooltip in bounds
        if (x + 320 > rect.width) x = e.clientX - rect.left - 330;
        if (y + 100 > rect.height) y = e.clientY - rect.top - 80;
        tooltip.style.left = x + "px";
        tooltip.style.top = y + "px";
      });

      svg.addEventListener("mouseout", function (e) {
        var t = e.target;
        if (t.tagName === "circle") {
          var related = e.relatedTarget;
          // Don't hide if moving to another circle with same node id
          if (related && related.tagName === "circle" &&
              related.getAttribute("data-node-id") === t.getAttribute("data-node-id")) return;
          currentHover = null;
          tooltip.style.display = "none";
        }
      });

      // Zoom with mouse wheel
      svg.addEventListener("wheel", function (e) {
        e.preventDefault();
        var rect = svg.getBoundingClientRect();
        // Mouse position as fraction of SVG element
        var mx = (e.clientX - rect.left) / rect.width;
        var my = (e.clientY - rect.top) / rect.height;
        // Convert to SVG coordinates
        var svgX = view.x + mx * view.w;
        var svgY = view.y + my * view.h;

        var zoomFactor = e.deltaY > 0 ? 1.15 : 0.87;
        var newW = view.w * zoomFactor;
        var newH = view.h * zoomFactor;

        // Clamp zoom
        var fullW = (data.max_x - data.min_x) + 2000;
        if (newW > fullW * 1.5) return; // don't zoom out too far
        if (newW < 500) return; // don't zoom in too far

        // Zoom centered on mouse position
        view.x = svgX - mx * newW;
        view.y = svgY - my * newH;
        view.w = newW;
        view.h = newH;
        updateViewBox();
      });

      // Pan with mouse drag
      var isDragging = false;
      var dragStart = { x: 0, y: 0 };
      var viewStart = { x: 0, y: 0 };

      svg.addEventListener("mousedown", function (e) {
        if (e.button !== 0) return;
        isDragging = true;
        svg.style.cursor = "grabbing";
        dragStart.x = e.clientX;
        dragStart.y = e.clientY;
        viewStart.x = view.x;
        viewStart.y = view.y;
        e.preventDefault();
      });

      window.addEventListener("mousemove", function (e) {
        if (!isDragging) return;
        var rect = svg.getBoundingClientRect();
        var dx = (e.clientX - dragStart.x) / rect.width * view.w;
        var dy = (e.clientY - dragStart.y) / rect.height * view.h;
        view.x = viewStart.x - dx;
        view.y = viewStart.y - dy;
        updateViewBox();
      });

      window.addEventListener("mouseup", function () {
        if (isDragging) {
          isDragging = false;
          svg.style.cursor = "grab";
        }
      });

      // Reset view button
      var resetBtn = document.createElement("button");
      resetBtn.textContent = "Reset View";
      resetBtn.style.cssText =
        "position:absolute;top:8px;right:8px;background:#AF8700;color:#000;border:none;" +
        "padding:4px 10px;border-radius:4px;font-size:12px;font-weight:bold;cursor:pointer;z-index:5;";
      resetBtn.addEventListener("click", function () {
        if (hasAlloc) {
          var pad = 800;
          view.x = allocMinX - pad;
          view.y = allocMinY - pad;
          view.w = (allocMaxX - allocMinX) + pad * 2;
          view.h = (allocMaxY - allocMinY) + pad * 2;
        }
        updateViewBox();
      });
      container.appendChild(resetBtn);

      // Help text
      var help = document.createElement("div");
      help.textContent = "Scroll to zoom · Drag to pan";
      help.style.cssText =
        "position:absolute;bottom:8px;left:8px;color:#666;font-size:11px;z-index:5;";
      container.appendChild(help);
    });
};
