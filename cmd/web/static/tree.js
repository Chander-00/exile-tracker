window.renderPassiveTree = function (containerId, allocatedNodeIds) {
  var container = document.getElementById(containerId);
  if (!container) return;

  var allocSet = {};
  for (var i = 0; i < allocatedNodeIds.length; i++) {
    allocSet[allocatedNodeIds[i]] = true;
  }

  fetch("/static/tree.json")
    .then(function (r) {
      return r.json();
    })
    .then(function (data) {
      var groups = data.groups;
      var nodes = data.nodes;
      var skillsPerOrbit = data.constants.skillsPerOrbit;
      var orbitRadii = data.constants.orbitRadii;
      var padding = 500;
      var minX = data.min_x - padding;
      var minY = data.min_y - padding;
      var vbW = data.max_x - data.min_x + padding * 2;
      var vbH = data.max_y - data.min_y + padding * 2;

      // Compute node positions
      var nodePos = {};
      var nodeKeys = Object.keys(nodes);
      for (var i = 0; i < nodeKeys.length; i++) {
        var id = nodeKeys[i];
        var node = nodes[id];
        var group = groups[node.g];
        if (!group) continue;
        var x, y;
        if (node.o === 0) {
          x = group.x;
          y = group.y;
        } else {
          var radius = orbitRadii[node.o] || 0;
          var total = skillsPerOrbit[node.o] || 1;
          var angle = (2 * Math.PI * node.oi) / total - Math.PI / 2;
          x = group.x + radius * Math.cos(angle);
          y = group.y + radius * Math.sin(angle);
        }
        nodePos[id] = { x: x, y: y };
      }

      var ns = "http://www.w3.org/2000/svg";
      var svg = document.createElementNS(ns, "svg");
      svg.setAttribute("width", "100%");
      svg.setAttribute("height", "100%");
      svg.setAttribute("viewBox", minX + " " + minY + " " + vbW + " " + vbH);
      svg.style.background = "transparent";

      // Draw connections
      for (var i = 0; i < nodeKeys.length; i++) {
        var id = nodeKeys[i];
        var node = nodes[id];
        var from = nodePos[id];
        if (!from || !node.out) continue;
        for (var j = 0; j < node.out.length; j++) {
          var targetId = node.out[j];
          var to = nodePos[targetId];
          if (!to) continue;
          var line = document.createElementNS(ns, "line");
          line.setAttribute("x1", from.x);
          line.setAttribute("y1", from.y);
          line.setAttribute("x2", to.x);
          line.setAttribute("y2", to.y);
          var bothAlloc = allocSet[id] && allocSet[targetId];
          line.setAttribute("stroke", bothAlloc ? "#d4a017" : "#3a3a3a");
          line.setAttribute("stroke-width", bothAlloc ? "2" : "1");
          svg.appendChild(line);
        }
      }

      // Node sizes by type
      var sizeMap = {
        notable: 5,
        keystone: 8,
        mastery: 4,
        jewel: 5,
        ascStart: 5,
      };

      // Draw nodes
      for (var i = 0; i < nodeKeys.length; i++) {
        var id = nodeKeys[i];
        var node = nodes[id];
        var pos = nodePos[id];
        if (!pos) continue;
        var r = sizeMap[node.t] || 2.5;
        var isAlloc = !!allocSet[id];
        var circle = document.createElementNS(ns, "circle");
        circle.setAttribute("cx", pos.x);
        circle.setAttribute("cy", pos.y);
        circle.setAttribute("r", r);
        circle.setAttribute("fill", isAlloc ? "#AF8700" : "#2a2a2a");
        if (isAlloc) {
          circle.setAttribute("stroke", "#d4a017");
          circle.setAttribute("stroke-width", "1.5");
        }
        circle.setAttribute("data-node-id", id);
        svg.appendChild(circle);
      }

      container.style.position = "relative";
      container.appendChild(svg);

      // Tooltip
      var tooltip = document.createElement("div");
      tooltip.style.cssText =
        "position:absolute;display:none;background:#1a1a2e;border:1px solid #AF8700;color:#ccc;padding:8px 12px;border-radius:4px;font-size:12px;pointer-events:none;z-index:10;max-width:300px;white-space:pre-wrap;";
      container.appendChild(tooltip);

      svg.addEventListener("mouseenter", function (e) {
        var t = e.target;
        if (t.tagName !== "circle") return;
        var nid = t.getAttribute("data-node-id");
        var node = nodes[nid];
        if (!node) return;
        var text = node.n || "";
        if (node.s && node.s.length > 0) {
          text += "\n" + node.s.join("\n");
        }
        tooltip.textContent = text;
        tooltip.style.display = "block";
      }, true);

      svg.addEventListener("mousemove", function (e) {
        if (tooltip.style.display === "none") return;
        var rect = container.getBoundingClientRect();
        var x = e.clientX - rect.left + 15;
        var y = e.clientY - rect.top + 15;
        tooltip.style.left = x + "px";
        tooltip.style.top = y + "px";
      }, true);

      svg.addEventListener("mouseleave", function (e) {
        if (e.target.tagName === "circle") {
          tooltip.style.display = "none";
        }
      }, true);
    });
};
