package models

import "encoding/json"

type POECharacterResponse struct {
	Name     string `json:"name"`
	Realm    string `json:"realm"`
	Class    string `json:"class"`
	League   string `json:"league"`
	Level    int    `json:"level"`
	Pinnable bool   `json:"pinnable"`
}

// PassiveSkillsResponse represents the response from the get-passive-skills endpoint
type PassiveSkillsResponse struct {
	Character           int                    `json:"character"`
	Ascendancy          int                    `json:"ascendancy"`
	AlternateAscendancy int                    `json:"alternate_ascendancy"`
	Hashes              []int                  `json:"hashes"`
	HashesEx            []int                  `json:"hashes_ex"`
	MasteryEffects      json.RawMessage        `json:"mastery_effects"`
	SkillOverrides      json.RawMessage        `json:"skill_overrides"`
	Items               []PassiveItem          `json:"items"`
	JewelData           json.RawMessage        `json:"jewel_data"`
}

// PassiveItem represents an item in the passive skills response
type PassiveItem struct {
	// Duplicated    bool           `json:"duplicated"`
	Verified      bool           `json:"verified"`
	W             int            `json:"w"`
	H             int            `json:"h"`
	Icon          string         `json:"icon"`
	League        string         `json:"league"`
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	TypeLine      string         `json:"typeLine"`
	BaseType      string         `json:"baseType"`
	Rarity        string         `json:"rarity"`
	Ilvl          int            `json:"ilvl"`
	Identified    bool           `json:"identified"`
	Properties    []ItemProperty `json:"properties,omitempty"`
	ExplicitMods  []string       `json:"explicitMods,omitempty"`
	DescrText     string         `json:"descrText"`
	FlavourText   []string       `json:"flavourText,omitempty"`
	FrameType     int            `json:"frameType"`
	X             int            `json:"x"`
	Y             int            `json:"y"`
	InventoryID   string         `json:"inventoryId"`
	Corrupted     bool           `json:"corrupted,omitempty"`
	ImplicitMods  []string       `json:"implicitMods,omitempty"`
	Fractured     bool           `json:"fractured,omitempty"`
	FracturedMods []string       `json:"fracturedMods,omitempty"`
	Requirements  []ItemProperty `json:"requirements,omitempty"`
	EnchantMods   []string       `json:"enchantMods,omitempty"`
}

// ItemProperty represents a property of an item
type ItemProperty struct {
	Name        string          `json:"name"`
	Values      [][]interface{} `json:"values"`
	DisplayMode int             `json:"displayMode"`
	Type        int             `json:"type,omitempty"`
	Suffix      string          `json:"suffix,omitempty"`
	Progress    float32         `json:"progress,omitempty"`
}

// JewelData represents jewel data in the passive skills response
type JewelData struct {
	Type         string         `json:"type"`
	Radius       int            `json:"radius,omitempty"`
	RadiusVisual string         `json:"radiusVisual,omitempty"`
	Subgraph     *JewelSubgraph `json:"subgraph,omitempty"`
}

// JewelSubgraph represents a subgraph within jewel data
type JewelSubgraph struct {
	Groups map[string]JewelGroup `json:"groups"`
	Nodes  map[string]JewelNode  `json:"nodes"`
}

// JewelGroup represents a group within a jewel subgraph
type JewelGroup struct {
	Proxy  string   `json:"proxy,omitempty"`
	Nodes  []string `json:"nodes"`
	X      float64  `json:"x"`
	Y      float64  `json:"y"`
	Orbits []int    `json:"orbits"`
}

// JewelNode represents a node within a jewel subgraph
type JewelNode struct {
	Skill          string          `json:"skill,omitempty"`
	Name           string          `json:"name,omitempty"`
	Icon           string          `json:"icon,omitempty"`
	Stats          []string        `json:"stats"`
	Group          string          `json:"group"`
	Orbit          int             `json:"orbit"`
	OrbitIndex     int             `json:"orbitIndex"`
	Out            []string        `json:"out"`
	In             []string        `json:"in"`
	IsJewelSocket  bool            `json:"isJewelSocket,omitempty"`
	IsMastery      bool            `json:"isMastery,omitempty"`
	IsNotable      bool            `json:"isNotable,omitempty"`
	ReminderText   []string        `json:"reminderText,omitempty"`
	ExpansionJewel *ExpansionJewel `json:"expansionJewel,omitempty"`
}

// ExpansionJewel represents expansion jewel data
type ExpansionJewel struct {
	Size   int    `json:"size"`
	Index  int    `json:"index"`
	Proxy  string `json:"proxy"`
	Parent string `json:"parent"`
}

// ItemsResponse represents the response from the get-items endpoint
type ItemsResponse struct {
	Items     []Item `json:"items"`
	Character struct {
		Name   string `json:"name"`
		Realm  string `json:"realm"`
		Class  string `json:"class"`
		League string `json:"league"`
		Level  int    `json:"level"`
	} `json:"character"`
}

// Item represents an item in the items response
type Item struct {
	Verified              bool            `json:"verified"`
	W                     int             `json:"w"`
	H                     int             `json:"h"`
	Icon                  string          `json:"icon"`
	League                string          `json:"league"`
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	TypeLine              string          `json:"typeLine"`
	BaseType              string          `json:"baseType"`
	Rarity                string          `json:"rarity"`
	Ilvl                  int             `json:"ilvl"`
	Identified            bool            `json:"identified"`
	Duplicated            bool            `json:"duplicated,omitempty"`
	Properties            []ItemProperty  `json:"properties,omitempty"`
	Requirements          []ItemProperty  `json:"requirements,omitempty"`
	ExplicitMods          []string        `json:"explicitMods,omitempty"`
	ImplicitMods          []string        `json:"implicitMods,omitempty"`
	DescrText             string          `json:"descrText,omitempty"`
	FlavourText           []string        `json:"flavourText,omitempty"`
	FrameType             int             `json:"frameType"`
	X                     int             `json:"x"`
	Y                     int             `json:"y"`
	InventoryID           string          `json:"inventoryId"`
	Corrupted             bool            `json:"corrupted,omitempty"`
	Fractured             bool            `json:"fractured,omitempty"`
	FracturedMods         []string        `json:"fracturedMods,omitempty"`
	EnchantMods           []string        `json:"enchantMods,omitempty"`
	CraftedMods           []string        `json:"craftedMods,omitempty"`
	UtilityMods           []string        `json:"utilityMods,omitempty"`
	Sockets               []Socket        `json:"sockets,omitempty"`
	SocketedItems         []SocketedItem  `json:"socketedItems,omitempty"`
	Influences            map[string]bool `json:"influences,omitempty"`
	Elder                 bool            `json:"elder,omitempty"`
	Shaper                bool            `json:"shaper,omitempty"`
	Searing               bool            `json:"searing,omitempty"`
	Tangled               bool            `json:"tangled,omitempty"`
	Split                 bool            `json:"split,omitempty"`
	IsRelic               bool            `json:"isRelic,omitempty"`
	FoilVariation         int             `json:"foilVariation,omitempty"`
	Hybrid                *HybridItem     `json:"hybrid,omitempty"`
	AdditionalProperties  []ItemProperty  `json:"additionalProperties,omitempty"`
	NextLevelRequirements []ItemProperty  `json:"nextLevelRequirements,omitempty"`
	Support               bool            `json:"support,omitempty"`
	Socket                int             `json:"socket,omitempty"`
	Colour                string          `json:"colour,omitempty"`
}

// Socket represents a socket on an item
type Socket struct {
	Group   int    `json:"group"`
	Attr    string `json:"attr"`
	SColour string `json:"sColour"`
}

// SocketedItem represents an item socketed in another item
type SocketedItem struct {
	Verified              bool           `json:"verified"`
	W                     int            `json:"w"`
	H                     int            `json:"h"`
	Icon                  string         `json:"icon"`
	Support               bool           `json:"support"`
	League                string         `json:"league"`
	ID                    string         `json:"id"`
	Name                  string         `json:"name"`
	TypeLine              string         `json:"typeLine"`
	BaseType              string         `json:"baseType"`
	Ilvl                  int            `json:"ilvl"`
	Identified            bool           `json:"identified"`
	Properties            []ItemProperty `json:"properties,omitempty"`
	Requirements          []ItemProperty `json:"requirements,omitempty"`
	ExplicitMods          []string       `json:"explicitMods,omitempty"`
	DescrText             string         `json:"descrText"`
	SecDescrText          string         `json:"secDescrText,omitempty"`
	FrameType             int            `json:"frameType"`
	Socket                int            `json:"socket"`
	Colour                string         `json:"colour"`
	Corrupted             bool           `json:"corrupted,omitempty"`
	Hybrid                *HybridItem    `json:"hybrid,omitempty"`
	AdditionalProperties  []ItemProperty `json:"additionalProperties,omitempty"`
	NextLevelRequirements []ItemProperty `json:"nextLevelRequirements,omitempty"`
}

// HybridItem represents hybrid item data (like Vaal gems)
type HybridItem struct {
	IsVaalGem    bool           `json:"isVaalGem,omitempty"`
	BaseTypeName string         `json:"baseTypeName,omitempty"`
	Properties   []ItemProperty `json:"properties,omitempty"`
	ExplicitMods []string       `json:"explicitMods,omitempty"`
	SecDescrText string         `json:"secDescrText,omitempty"`
}
