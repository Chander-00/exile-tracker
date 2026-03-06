// shrinktree reads GGG's passive skill tree data.json and outputs a minimal
// tree.json suitable for rendering an SVG tree on the web frontend.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// ---------- input types (only the fields we care about) ----------

type RawData struct {
	Groups    map[string]RawGroup `json:"groups"`
	Nodes     map[string]RawNode  `json:"nodes"`
	Constants json.RawMessage     `json:"constants"`
	MinX      json.Number         `json:"min_x"`
	MinY      json.Number         `json:"min_y"`
	MaxX      json.Number         `json:"max_x"`
	MaxY      json.Number         `json:"max_y"`
}

type RawGroup struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type RawNode struct {
	Skill                  int      `json:"skill"`
	Name                   string   `json:"name"`
	Group                  int      `json:"group"`
	Orbit                  int      `json:"orbit"`
	OrbitIndex             int      `json:"orbitIndex"`
	Out                    []string `json:"out"`
	Stats                  []string `json:"stats"`
	AscendancyName         string   `json:"ascendancyName"`
	ClassStartIndex        *int     `json:"classStartIndex"`
	IsNotable              bool     `json:"isNotable"`
	IsKeystone             bool     `json:"isKeystone"`
	IsMastery              bool     `json:"isMastery"`
	IsJewelSocket          bool     `json:"isJewelSocket"`
	IsAscendancyStart      bool     `json:"isAscendancyStart"`
	IsBlighted             bool     `json:"isBlighted"`
	IsMultipleChoiceOption bool     `json:"isMultipleChoiceOption"`
}

// ---------- output types (short keys, omitempty) ----------

type OutData struct {
	Groups    map[string]OutGroup `json:"groups"`
	Nodes     map[string]OutNode  `json:"nodes"`
	Constants OutConstants        `json:"constants"`
	MinX      json.Number         `json:"min_x"`
	MinY      json.Number         `json:"min_y"`
	MaxX      json.Number         `json:"max_x"`
	MaxY      json.Number         `json:"max_y"`
}

type OutGroup struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type OutNode struct {
	G   int      `json:"g"`
	O   int      `json:"o"`
	OI  int      `json:"oi"`
	Out []string `json:"out,omitempty"`
	N   string   `json:"n,omitempty"`
	S   []string `json:"s,omitempty"`
	T   string   `json:"t,omitempty"`
	A   string   `json:"a,omitempty"`
}

type OutConstants struct {
	SkillsPerOrbit []int `json:"skillsPerOrbit"`
	OrbitRadii     []int `json:"orbitRadii"`
}

type RawConstants struct {
	SkillsPerOrbit []int `json:"skillsPerOrbit"`
	OrbitRadii     []int `json:"orbitRadii"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: shrinktree <data.json> [output.json]")
		os.Exit(1)
	}

	inPath := os.Args[1]
	outPath := "cmd/web/static/tree.json"
	if len(os.Args) >= 3 {
		outPath = os.Args[2]
	}

	// ---- read ----
	raw, err := os.ReadFile(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", inPath, err)
		os.Exit(1)
	}
	fmt.Printf("input:  %s (%d bytes)\n", inPath, len(raw))

	var data RawData
	if err := json.Unmarshal(raw, &data); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// ---- parse constants ----
	var rc RawConstants
	if err := json.Unmarshal(data.Constants, &rc); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing constants: %v\n", err)
		os.Exit(1)
	}

	// ---- filter nodes ----
	outNodes := make(map[string]OutNode, len(data.Nodes))
	skipped := 0
	for id, n := range data.Nodes {
		// skip root
		if id == "root" {
			skipped++
			continue
		}
		// skip class start nodes
		if n.ClassStartIndex != nil {
			skipped++
			continue
		}
		// skip blighted
		if n.IsBlighted {
			skipped++
			continue
		}
		// skip non-ascendancy multiple choice options
		if n.IsMultipleChoiceOption && n.AscendancyName == "" {
			skipped++
			continue
		}
		// skip unnamed non-ascendancy-start nodes
		if n.Name == "" && !n.IsAscendancyStart {
			skipped++
			continue
		}

		on := OutNode{
			G:  n.Group,
			O:  n.Orbit,
			OI: n.OrbitIndex,
		}
		if len(n.Out) > 0 {
			on.Out = n.Out
		}
		if n.Name != "" {
			on.N = n.Name
		}
		if len(n.Stats) > 0 {
			on.S = n.Stats
		}
		if n.AscendancyName != "" {
			on.A = n.AscendancyName
		}

		// determine type
		switch {
		case n.IsKeystone:
			on.T = "keystone"
		case n.IsNotable:
			on.T = "notable"
		case n.IsMastery:
			on.T = "mastery"
		case n.IsJewelSocket:
			on.T = "jewel"
		case n.IsAscendancyStart:
			on.T = "ascStart"
		}

		outNodes[id] = on
	}

	// ---- keep only referenced groups ----
	usedGroups := make(map[string]bool, len(outNodes))
	for _, n := range outNodes {
		usedGroups[fmt.Sprintf("%d", n.G)] = true
	}
	outGroups := make(map[string]OutGroup, len(usedGroups))
	for gid, g := range data.Groups {
		if usedGroups[gid] {
			outGroups[gid] = OutGroup{X: g.X, Y: g.Y}
		}
	}

	// ---- assemble output ----
	out := OutData{
		Groups: outGroups,
		Nodes:  outNodes,
		Constants: OutConstants{
			SkillsPerOrbit: rc.SkillsPerOrbit,
			OrbitRadii:     rc.OrbitRadii,
		},
		MinX: data.MinX,
		MinY: data.MinY,
		MaxX: data.MaxX,
		MaxY: data.MaxY,
	}

	outJSON, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling output: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, outJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("output: %s (%d bytes)\n", outPath, len(outJSON))
	fmt.Printf("nodes:  %d kept, %d skipped (of %d)\n", len(outNodes), skipped, len(data.Nodes))
	fmt.Printf("groups: %d kept (of %d)\n", len(outGroups), len(data.Groups))
	fmt.Printf("ratio:  %.1f%% of original\n", float64(len(outJSON))/float64(len(raw))*100)
}
