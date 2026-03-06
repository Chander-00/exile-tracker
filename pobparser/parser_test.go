package pobparser

import (
	"os"
	"testing"
)

func TestDecodeAndSummarize(t *testing.T) {
	data, err := os.ReadFile("/tmp/pob_test.txt")
	if err != nil {
		t.Skip("no test file at /tmp/pob_test.txt")
	}

	code := string(data)
	code = code[:len(code)-1] // trim newline

	pob, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	summary := pob.Summarize()

	if summary.Class == "" {
		t.Error("expected class to be set")
	}
	if summary.Level == 0 {
		t.Error("expected level > 0")
	}

	t.Logf("Class: %s, Ascendancy: %s, Level: %d", summary.Class, summary.Ascendancy, summary.Level)

	// Stats
	for _, key := range []string{"Life", "Mana", "EnergyShield", "Evasion", "Armour", "CombinedDPS"} {
		if val, ok := summary.Stats[key]; ok {
			t.Logf("  %s: %.0f", key, val)
		}
	}

	// Resistances
	for _, key := range []string{"FireResist", "ColdResist", "LightningResist", "ChaosResist"} {
		if val, ok := summary.Stats[key]; ok {
			t.Logf("  %s: %.0f", key, val)
		}
	}

	// Items
	t.Logf("Items (%d):", len(summary.Items))
	for _, item := range summary.Items {
		t.Logf("  [%s] %s - %s (%s)", item.Slot, item.Name, item.BaseName, item.Rarity)
	}

	// Skills
	t.Logf("Skill Groups (%d):", len(summary.SkillGroups))
	for _, sg := range summary.SkillGroups {
		t.Logf("  [%s] %v", sg.Slot, sg.Gems)
	}

	// Tree
	t.Logf("Tree URL: %s", summary.TreeURL)
	t.Logf("Node Count: %d", summary.NodeCount)
	if summary.TreeURL == "" {
		t.Error("expected TreeURL to be set")
	}
	if summary.NodeCount == 0 {
		t.Error("expected NodeCount > 0")
	}
}
