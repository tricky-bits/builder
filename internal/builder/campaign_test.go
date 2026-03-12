package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCampaign_Validate(t *testing.T) {
	tests := []struct {
		name        string
		campaign    *Campaign
		wantErr     bool
		errContains string
	}{
		{
			name: "valid campaign",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{
					Title:    "My Campaign",
					Category: "cryptography",
					Slug:     "my-campaign",
				},
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Start: true}},
				},
			},
		},
		{
			name: "missing title",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{
					Category: "cryptography",
					Slug:     "my-campaign",
				},
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Start: true}},
				},
			},
			wantErr:     true,
			errContains: "title is required",
		},
		{
			name: "missing category",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{
					Title: "My Campaign",
					Slug:  "my-campaign",
				},
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Start: true}},
				},
			},
			wantErr:     true,
			errContains: "category is required",
		},
		{
			name: "no start stage",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{
					Title:    "My Campaign",
					Category: "cryptography",
					Slug:     "my-campaign",
				},
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Start: false}},
				},
			},
			wantErr:     true,
			errContains: "no start stage found",
		},
		{
			name: "multiple start stages",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{
					Title:    "My Campaign",
					Category: "cryptography",
					Slug:     "my-campaign",
				},
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Start: true}},
					"stage-02": {Frontmatter: StageFrontmatter{Start: true}},
				},
			},
			wantErr:     true,
			errContains: "multiple start stages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.campaign.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCampaign_ValidateLinearChain(t *testing.T) {
	tests := []struct {
		name        string
		campaign    *Campaign
		wantErr     bool
		errContains string
	}{
		{
			name: "valid single stage chain",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{Slug: "test"},
				StartSlug:   "stage-01",
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Next: ""}},
				},
			},
		},
		{
			name: "valid multi-stage chain",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{Slug: "test"},
				StartSlug:   "stage-01",
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Next: "stage-02"}},
					"stage-02": {Frontmatter: StageFrontmatter{Next: "stage-03"}},
					"stage-03": {Frontmatter: StageFrontmatter{Next: ""}},
				},
			},
		},
		{
			name: "cycle detection",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{Slug: "test"},
				StartSlug:   "stage-01",
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Next: "stage-02"}},
					"stage-02": {Frontmatter: StageFrontmatter{Next: "stage-01"}},
				},
			},
			wantErr:     true,
			errContains: "cycle detected",
		},
		{
			name: "orphaned stage",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{Slug: "test"},
				StartSlug:   "stage-01",
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Next: ""}},
					"stage-02": {Frontmatter: StageFrontmatter{Next: ""}},
				},
			},
			wantErr:     true,
			errContains: "orphaned stages",
		},
		{
			name: "next references missing stage",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{Slug: "test"},
				StartSlug:   "stage-01",
				Stages: map[string]*Stage{
					"stage-01": {Frontmatter: StageFrontmatter{Next: "does-not-exist"}},
				},
			},
			wantErr:     true,
			errContains: "stage not found",
		},
		{
			name: "no start slug",
			campaign: &Campaign{
				Frontmatter: CampaignFrontmatter{Slug: "test"},
				StartSlug:   "",
				Stages:      map[string]*Stage{},
			},
			wantErr:     true,
			errContains: "no start stage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.campaign.ValidateLinearChain()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadCampaign(t *testing.T) {
	t.Run("valid campaign loads successfully", func(t *testing.T) {
		campaign, err := LoadCampaign("../../testdata/campaigns/valid-campaign")
		require.NoError(t, err)
		require.NotNil(t, campaign)

		assert.Equal(t, "Valid Campaign", campaign.Frontmatter.Title)
		assert.Equal(t, "cryptography", campaign.Frontmatter.Category)
		assert.Equal(t, "valid-campaign", campaign.Frontmatter.Slug)
		assert.Equal(t, "stage-01", campaign.StartSlug)
		assert.Len(t, campaign.Stages, 2)
		assert.Contains(t, campaign.Stages, "stage-01")
		assert.Contains(t, campaign.Stages, "stage-02")
	})

	t.Run("missing campaign.md returns descriptive error", func(t *testing.T) {
		_, err := LoadCampaign("../../testdata/campaigns/does-not-exist")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "campaign.md")
	})

	t.Run("campaign with no start stage returns error", func(t *testing.T) {
		_, err := LoadCampaign("../../testdata/campaigns/missing-start")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no start stage found")
	})
}
