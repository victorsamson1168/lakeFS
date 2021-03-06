package retention_test

import (
	"testing"

	"github.com/go-test/deep"

	"github.com/treeverse/lakefs/api/gen/models"
	"github.com/treeverse/lakefs/catalog"
	"github.com/treeverse/lakefs/retention"
)

func TestParseTimePeriod(t *testing.T) {
	cases := []struct {
		Input  models.TimePeriod
		Output catalog.TimePeriodHours
		Error  bool
	}{
		{Input: models.TimePeriod{}, Error: true},
		{Input: models.TimePeriod{Days: 2}, Output: catalog.TimePeriodHours(2 * 24)},
		{Input: models.TimePeriod{Weeks: 5}, Output: catalog.TimePeriodHours(5 * 24 * 7)},
		{Input: models.TimePeriod{Weeks: 1, Days: 3}, Output: catalog.TimePeriodHours(10 * 24)},
	}
	for _, c := range cases {
		got, err := retention.ParseTimePeriod(c.Input)
		if !c.Error {
			if err != nil {
				t.Errorf("unexpected error parsing %#v: \"%s\"", c.Input, err)
				continue
			}
			if c.Output != got {
				t.Errorf("expected %#v to return %d, got %d hours", c.Input, c.Output, got)
			}
		} else {
			if err == nil {
				t.Errorf("expected error parsing %#v, got %d hours", c.Input, got)
			}
		}
	}
}

func TestUnparseTimePeriod(t *testing.T) {
	cases := []struct {
		Input  catalog.TimePeriodHours
		Output models.TimePeriod
	}{
		{Output: models.TimePeriod{Days: 2}, Input: catalog.TimePeriodHours(2*24 - 20)},
		{Output: models.TimePeriod{Weeks: 5}, Input: catalog.TimePeriodHours(5 * 24 * 7)},
		{Output: models.TimePeriod{Weeks: 1, Days: 3}, Input: catalog.TimePeriodHours(10 * 24)},
	}
	for _, c := range cases {
		got := retention.RenderTimePeriod(c.Input)
		if c.Output != *got {
			t.Errorf("expected %d to return %#v, got %#v", c.Input, c.Output, got)
		}
	}
}

func TestParseExpiration(t *testing.T) {
	hours := func(h int) *catalog.TimePeriodHours {
		ret := catalog.TimePeriodHours(h)
		return &ret
	}
	cases := []struct {
		Input  models.RetentionPolicyRuleExpiration
		Output catalog.Expiration
		Error  bool
	}{
		{
			Input: models.RetentionPolicyRuleExpiration{},
			Error: true,
		}, {
			Input:  models.RetentionPolicyRuleExpiration{All: &models.TimePeriod{Days: 3}},
			Output: catalog.Expiration{All: hours(3 * 24)},
		}, {
			Input:  models.RetentionPolicyRuleExpiration{Noncurrent: &models.TimePeriod{Days: 3}},
			Output: catalog.Expiration{Noncurrent: hours(3 * 24)},
		}, {
			Input:  models.RetentionPolicyRuleExpiration{Uncommitted: &models.TimePeriod{Days: 3}},
			Output: catalog.Expiration{Uncommitted: hours(3 * 24)},
		}, {
			Input:  models.RetentionPolicyRuleExpiration{All: &models.TimePeriod{Days: 3}},
			Output: catalog.Expiration{All: hours(3 * 24)},
		}, {
			Input: models.RetentionPolicyRuleExpiration{
				All:         &models.TimePeriod{Days: 3},
				Uncommitted: &models.TimePeriod{Days: 1},
			},
			Output: catalog.Expiration{All: hours(3 * 24), Uncommitted: hours(24)},
		},
	}

	for _, c := range cases {
		got, err := retention.ParseExpiration(c.Input)
		if !c.Error {
			if err != nil {
				t.Errorf("unexpected error parsing %#v: \"%s\"", c.Input, err)
				continue
			}
			diff := deep.Equal(got, &c.Output)
			if diff != nil {
				t.Errorf("%#v: difference %s (expected %#v, got %#v)", c.Input, diff, c.Output, got)
			}
		} else {
			if err == nil {
				t.Errorf("expected error parsing %#v, got %#v", c.Input, got)
			}
		}
	}
}

func TestUnParseExpiration(t *testing.T) {
	hours := func(h int) *catalog.TimePeriodHours {
		ret := catalog.TimePeriodHours(h)
		return &ret
	}
	cases := []struct {
		Output models.RetentionPolicyRuleExpiration
		Input  catalog.Expiration
	}{
		{
			Output: models.RetentionPolicyRuleExpiration{All: &models.TimePeriod{Days: 3}},
			Input:  catalog.Expiration{All: hours(3 * 24)},
		}, {
			Output: models.RetentionPolicyRuleExpiration{Noncurrent: &models.TimePeriod{Days: 3}},
			Input:  catalog.Expiration{Noncurrent: hours(3 * 24)},
		}, {
			Output: models.RetentionPolicyRuleExpiration{Uncommitted: &models.TimePeriod{Days: 3}},
			Input:  catalog.Expiration{Uncommitted: hours(3 * 24)},
		}, {
			Output: models.RetentionPolicyRuleExpiration{All: &models.TimePeriod{Days: 3}},
			Input:  catalog.Expiration{All: hours(3 * 24)},
		}, {
			Output: models.RetentionPolicyRuleExpiration{
				All:         &models.TimePeriod{Days: 3},
				Uncommitted: &models.TimePeriod{Days: 1},
			},
			Input: catalog.Expiration{All: hours(3 * 24), Uncommitted: hours(24)},
		},
	}

	for _, c := range cases {
		got := retention.RenderExpiration(&c.Input)
		diff := deep.Equal(got, &c.Output)
		if diff != nil {
			t.Errorf("%#v: difference %s (expected %#v, got %#v)", c.Input, diff, c.Output, got)
		}
	}
}

// ParseRule just parses all fields, skip testing it.

func TestParsePolicy(t *testing.T) {
	enabled := "enabled"
	disabled := "disabled"
	pathA := "/bucket/a"
	pathB := "/bucket/b"
	modelA := &models.RetentionPolicyRule{
		Status:     &enabled,
		Filter:     &models.RetentionPolicyRuleFilter{Prefix: pathA},
		Expiration: &models.RetentionPolicyRuleExpiration{All: &models.TimePeriod{Days: 1}},
	}
	modelB := &models.RetentionPolicyRule{
		Status:     &disabled,
		Filter:     &models.RetentionPolicyRuleFilter{Prefix: pathB},
		Expiration: &models.RetentionPolicyRuleExpiration{Uncommitted: &models.TimePeriod{Days: 1}},
	}
	modelFail := &models.RetentionPolicyRule{Status: &enabled}

	cases := []struct {
		Input        models.RetentionPolicy
		OutputPrefix []string // (enough to ID the rule, other parsing checked elsewhere)
		Error        bool
	}{
		{Input: models.RetentionPolicy{Rules: []*models.RetentionPolicyRule{modelFail}}, Error: true},
		{Input: models.RetentionPolicy{Rules: []*models.RetentionPolicyRule{modelA, modelFail, modelB}}, Error: true},
		{Input: models.RetentionPolicy{Rules: []*models.RetentionPolicyRule{modelA}}, OutputPrefix: []string{pathA}},
		{Input: models.RetentionPolicy{Rules: []*models.RetentionPolicyRule{modelA, modelB}}, OutputPrefix: []string{pathA, pathB}},
	}
	for _, c := range cases {
		got, err := retention.ParsePolicy(c.Input)
		if !c.Error {
			if err != nil {
				t.Errorf("unexpected error parsing %#v: \"%s\"", c.Input, err)
				continue
			}
			gotPrefix := make([]string, 0, len(got.Rules))
			for _, rule := range got.Rules {
				gotPrefix = append(gotPrefix, rule.FilterPrefix)
			}
			diff := deep.Equal(c.OutputPrefix, gotPrefix)
			if diff != nil {
				t.Errorf("%#v: difference %s (expected prefixes %v, got %v", c.Input, diff, c.OutputPrefix, gotPrefix)
			}
		} else {
			if err == nil {
				t.Errorf("expected error parsing %#v, got %#v", c.Input, got)
			}
		}
	}
}

func TestUnparsePolicy(t *testing.T) {
	pathA := "/bucket/a"
	pathB := "/bucket/b"
	day := catalog.TimePeriodHours(24)
	ruleA := catalog.Rule{
		Enabled:      true,
		FilterPrefix: pathA,
		Expiration:   catalog.Expiration{All: &day},
	}
	ruleB := catalog.Rule{
		Enabled:      false,
		FilterPrefix: pathB,
		Expiration:   catalog.Expiration{Uncommitted: &day},
	}

	cases := []struct {
		Input        catalog.Policy
		OutputPrefix []string // (enough to ID the rule, other parsing checked elsewhere)
	}{
		{Input: catalog.Policy{Rules: []catalog.Rule{ruleA}}, OutputPrefix: []string{pathA}},
		{Input: catalog.Policy{Rules: []catalog.Rule{ruleA, ruleB}}, OutputPrefix: []string{pathA, pathB}},
	}
	for _, c := range cases {
		got := retention.RenderPolicy(&c.Input)
		gotPrefix := make([]string, 0, len(got.Rules))
		for _, rule := range got.Rules {
			gotPrefix = append(gotPrefix, rule.Filter.Prefix)
		}
		diff := deep.Equal(c.OutputPrefix, gotPrefix)
		if diff != nil {
			t.Errorf("%#v: difference %s (expected prefixes %v, got %v", c.Input, diff, c.OutputPrefix, gotPrefix)
		}
	}
}
