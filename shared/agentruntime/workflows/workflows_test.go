package workflows

import "testing"

func TestDefaultTemplatesValidate(t *testing.T) {
	for _, template := range DefaultTemplates() {
		t.Run(template.ID, func(t *testing.T) {
			if err := Validate(template); err != nil {
				t.Fatalf("Validate error: %v", err)
			}
		})
	}
}

func TestLibraryListsTemplatesSorted(t *testing.T) {
	library := DefaultLibrary()
	templates := library.List()
	if len(templates) < 3 {
		t.Fatalf("List returned %d templates, want built-ins", len(templates))
	}
	for i := 1; i < len(templates); i++ {
		if templates[i-1].ID > templates[i].ID {
			t.Fatalf("templates not sorted: %q before %q", templates[i-1].ID, templates[i].ID)
		}
	}
}

func TestFeatureTemplateHasReviewableGatesAndValidationLoop(t *testing.T) {
	template, ok := DefaultLibrary().Get("feature")
	if !ok {
		t.Fatal("feature template missing")
	}
	var planning, validation, pr Stage
	for _, stage := range template.Stages {
		switch stage.ID {
		case "planning":
			planning = stage
		case "validation":
			validation = stage
		case "pull-request":
			pr = stage
		}
	}
	if planning.Gate != GateHuman {
		t.Fatalf("planning gate = %q, want human", planning.Gate)
	}
	if pr.Gate != GateHuman {
		t.Fatalf("pull-request gate = %q, want human", pr.Gate)
	}
	if len(validation.OnFailure) != 1 || validation.OnFailure[0] != "implementation" {
		t.Fatalf("validation failure path = %#v, want implementation loop", validation.OnFailure)
	}
}

func TestResearchTemplateHasNoImplementationOrPR(t *testing.T) {
	template, ok := DefaultLibrary().Get("research")
	if !ok {
		t.Fatal("research template missing")
	}
	for _, stage := range template.Stages {
		if stage.Kind == StageImplementation || stage.Kind == StagePullRequest {
			t.Fatalf("research template should not include %q stage", stage.Kind)
		}
	}
}

func TestValidateRejectsUnknownStageReference(t *testing.T) {
	template := Template{
		ID:   "bad",
		Name: "Bad",
		Stages: []Stage{
			{ID: "one", Kind: StageIntake, Name: "One", Gate: GateNone, Next: []string{"missing"}},
		},
	}
	if err := Validate(template); err == nil {
		t.Fatal("Validate should reject unknown stage reference")
	}
}
