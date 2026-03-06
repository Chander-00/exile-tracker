package pobparser

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type PathOfBuilding struct {
	Build  Build  `xml:"Build"`
	Items  Items  `xml:"Items"`
	Skills Skills `xml:"Skills"`
	Tree   Tree   `xml:"Tree"`
}

type Build struct {
	ClassName      string       `xml:"className,attr"`
	Level          int          `xml:"level,attr"`
	AscendClass    string       `xml:"ascendClassName,attr"`
	MainSocketGrp  string       `xml:"mainSocketGroup,attr"`
	PlayerStats    []PlayerStat `xml:"PlayerStat"`
}

type PlayerStat struct {
	Stat  string `xml:"stat,attr"`
	Value string `xml:"value,attr"`
}

type Items struct {
	RawItems []RawItem `xml:"Item"`
	Slots    []Slot    `xml:"ItemSet>Slot"`
}

type RawItem struct {
	ID      int    `xml:"id,attr"`
	Content string `xml:",chardata"`
}

type Slot struct {
	Name   string `xml:"name,attr"`
	ItemID int    `xml:"itemId,attr"`
}

type Skills struct {
	SkillSets []SkillSet `xml:"SkillSet"`
}

type SkillSet struct {
	ID     string  `xml:"id,attr"`
	Skills []Skill `xml:"Skill"`
}

type Skill struct {
	Slot    string `xml:"slot,attr"`
	Label   string `xml:"label,attr"`
	Enabled string `xml:"enabled,attr"`
	Gems    []Gem  `xml:"Gem"`
}

type Gem struct {
	NameSpec string `xml:"nameSpec,attr"`
	Level    int    `xml:"level,attr"`
	Quality  int    `xml:"quality,attr"`
	Enabled  string `xml:"enabled,attr"`
	SkillID  string `xml:"skillId,attr"`
}

type Tree struct {
	ActiveSpec string `xml:"activeSpec,attr"`
	Specs      []Spec `xml:"Spec"`
}

type Spec struct {
	ClassID       string `xml:"classId,attr"`
	AscendClassID string `xml:"ascendClassId,attr"`
	TreeVersion   string `xml:"treeVersion,attr"`
	Nodes         string `xml:"nodes,attr"`
	URL           string `xml:"URL"`
}

// Parsed high-level build summary for display.
type BuildSummary struct {
	Class       string
	Ascendancy  string
	Level       int
	Stats       map[string]float64
	Items       []ParsedItem
	SkillGroups []SkillGroup
	TreeURL     string
	NodeCount   int
}

type ParsedItem struct {
	Slot     string
	Rarity   string
	Name     string
	BaseName string
}

type SkillGroup struct {
	Slot string
	Gems []string
}

func Decode(pobCode string) (*PathOfBuilding, error) {
	decoded, err := base64.URLEncoding.DecodeString(pobCode)
	if err != nil {
		// Try with padding
		padded := pobCode + strings.Repeat("=", (4-len(pobCode)%4)%4)
		decoded, err = base64.URLEncoding.DecodeString(padded)
		if err != nil {
			return nil, fmt.Errorf("base64 decode failed: %w", err)
		}
	}

	r, err := zlib.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("zlib reader failed: %w", err)
	}
	defer r.Close()

	xmlData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("zlib decompress failed: %w", err)
	}

	var pob PathOfBuilding
	if err := xml.Unmarshal(xmlData, &pob); err != nil {
		return nil, fmt.Errorf("xml parse failed: %w", err)
	}

	return &pob, nil
}

func (pob *PathOfBuilding) Summarize() BuildSummary {
	summary := BuildSummary{
		Class:      pob.Build.ClassName,
		Ascendancy: pob.Build.AscendClass,
		Level:      pob.Build.Level,
		Stats:      make(map[string]float64),
	}

	for _, ps := range pob.Build.PlayerStats {
		val, err := strconv.ParseFloat(ps.Value, 64)
		if err == nil {
			summary.Stats[ps.Stat] = val
		}
	}

	// Build item ID -> parsed item map
	itemMap := make(map[int]ParsedItem)
	for _, raw := range pob.Items.RawItems {
		itemMap[raw.ID] = parseRawItem(raw.Content)
	}

	// Map slots to items
	for _, slot := range pob.Items.Slots {
		if slot.ItemID == 0 {
			continue
		}
		if strings.Contains(slot.Name, "Abyssal") || strings.Contains(slot.Name, "Swap") {
			continue
		}
		if item, ok := itemMap[slot.ItemID]; ok {
			item.Slot = slot.Name
			summary.Items = append(summary.Items, item)
		}
	}

	// Extract tree info
	if len(pob.Tree.Specs) > 0 {
		spec := pob.Tree.Specs[0]
		summary.TreeURL = strings.TrimSpace(spec.URL)
		if spec.Nodes != "" {
			summary.NodeCount = len(strings.Split(spec.Nodes, ","))
		}
	}

	// Parse skill groups
	for _, ss := range pob.Skills.SkillSets {
		for _, skill := range ss.Skills {
			if skill.Enabled != "true" || len(skill.Gems) == 0 {
				continue
			}
			var gems []string
			for _, gem := range skill.Gems {
				if gem.Enabled != "true" {
					continue
				}
				gems = append(gems, gem.NameSpec)
			}
			if len(gems) > 0 {
				summary.SkillGroups = append(summary.SkillGroups, SkillGroup{
					Slot: skill.Slot,
					Gems: gems,
				})
			}
		}
	}

	return summary
}

func parseRawItem(content string) ParsedItem {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	item := ParsedItem{}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Rarity:") {
			item.Rarity = strings.TrimSpace(strings.TrimPrefix(line, "Rarity:"))
			// Next line(s) are name and optionally base type
			if i+1 < len(lines) {
				item.Name = strings.TrimSpace(lines[i+1])
			}
			if i+2 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+2])
				// If it's not a stat line, it's the base type
				if !strings.Contains(nextLine, ":") || item.Rarity == "UNIQUE" || item.Rarity == "RARE" {
					item.BaseName = nextLine
				}
			}
			break
		}
	}

	return item
}
