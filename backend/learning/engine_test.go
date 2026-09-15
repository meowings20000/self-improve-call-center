package learning

import (
	"regexp"
	"testing"
)

func TestNewSessionStartsWithPlayableExerciseWithoutAnswer(t *testing.T) {
	engine := NewEngine()
	session := engine.NewSession()

	if session.ID == "" {
		t.Fatal("expected a session ID")
	}
	if session.Hearts != 5 || session.XP != 0 || session.Answered != 0 {
		t.Fatalf("unexpected initial progress: %+v", session)
	}
	exercise, err := engine.Next(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if exercise.ID == "" || exercise.Prompt == "" || exercise.Skill == "" || len(exercise.Choices) < 2 {
		t.Fatalf("expected a playable exercise, got %+v", exercise)
	}
}

func TestNewSessionReturnsIndependentSkillsSnapshot(t *testing.T) {
	engine := NewEngine()
	session := engine.NewSession()
	session.Skills["injected"] = 999

	progress, err := engine.Progress(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := progress.Skills["injected"]; exists {
		t.Fatal("mutating new-session response changed engine skills")
	}
}

func TestNewSessionUsesOpaqueRandomID(t *testing.T) {
	engine := NewEngine()
	first := engine.NewSession()
	second := engine.NewSession()

	opaqueID := regexp.MustCompile(`^[a-f0-9]{32}$`)
	if !opaqueID.MatchString(first.ID) || !opaqueID.MatchString(second.ID) {
		t.Fatalf("session IDs must be 128-bit hexadecimal tokens, got %q and %q", first.ID, second.ID)
	}
	if first.ID == second.ID {
		t.Fatalf("session IDs must be unique, both were %q", first.ID)
	}
}

func TestNewSessionEvictsOldestSessionAtCapacity(t *testing.T) {
	engine := newEngine(2)
	oldest := engine.NewSession()
	engine.NewSession()
	engine.NewSession()

	if _, err := engine.Progress(oldest.ID); err == nil {
		t.Fatal("expected the oldest session to be evicted at capacity")
	}
	if got := len(engine.sessions); got != 2 {
		t.Fatalf("session count = %d, want 2", got)
	}
}

func TestCorrectAnswerGivesImmediateFeedbackAndXP(t *testing.T) {
	engine := NewEngine()
	session := engine.NewSession()
	exercise, _ := engine.Next(session.ID)

	result, err := engine.Check(session.ID, exercise.ID, "She walks to work every day.")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Correct || result.XPEarned != 10 || result.TotalXP != 10 {
		t.Fatalf("unexpected grade: %+v", result)
	}
	if result.Feedback == "" || result.Explanation == "" {
		t.Fatal("expected immediate explanatory feedback")
	}
}

func TestCheckRejectsExerciseNotIssuedToSession(t *testing.T) {
	engine := NewEngine()
	owner := engine.NewSession()
	exercise, _ := engine.Next(owner.ID)
	other := engine.NewSession()

	if _, err := engine.Check(other.ID, exercise.ID, exercise.answer); err == nil {
		t.Fatal("expected an exercise issued to another session to be rejected")
	}
	progress, _ := engine.Progress(other.ID)
	if progress.Answered != 0 || progress.XP != 0 || progress.Hearts != 5 {
		t.Fatalf("unissued answer changed progress: %+v", progress)
	}
}

func TestCheckRejectsRepeatedCorrectAnswerWithoutAwardingXP(t *testing.T) {
	engine := NewEngine()
	session := engine.NewSession()
	exercise, _ := engine.Next(session.ID)
	engine.Check(session.ID, exercise.ID, "She walks to work every day.")

	if _, err := engine.Check(session.ID, exercise.ID, "She walks to work every day."); err == nil {
		t.Fatal("expected a completed exercise to be rejected")
	}
	progress, _ := engine.Progress(session.ID)
	if progress.Answered != 1 || progress.Correct != 1 || progress.XP != 10 {
		t.Fatalf("repeated correct answer changed progress: %+v", progress)
	}
}

func TestCheckAllowsCorrectRetryAfterWrongAnswer(t *testing.T) {
	engine := NewEngine()
	session := engine.NewSession()
	exercise, _ := engine.Next(session.ID)
	engine.Check(session.ID, exercise.ID, "wrong")

	grade, err := engine.Check(session.ID, exercise.ID, exercise.answer)
	if err != nil {
		t.Fatalf("correct retry after a wrong answer was rejected: %v", err)
	}
	if !grade.Correct || grade.XPEarned != 10 || grade.TotalXP != 10 || grade.Hearts != 4 {
		t.Fatalf("unexpected retry grade: %+v", grade)
	}
}

func TestEmptyAnswerIsRejectedWithoutChangingProgress(t *testing.T) {
	engine := NewEngine()
	session := engine.NewSession()
	exercise, _ := engine.Next(session.ID)

	_, err := engine.Check(session.ID, exercise.ID, "   ")
	if err == nil {
		t.Fatal("expected empty answer validation error")
	}
	progress, _ := engine.Progress(session.ID)
	if progress.Answered != 0 || progress.Hearts != 5 {
		t.Fatalf("invalid input changed progress: %+v", progress)
	}
}

func TestProgressReturnsIndependentSkillsSnapshot(t *testing.T) {
	engine := NewEngine()
	session := engine.NewSession()
	exercise, _ := engine.Next(session.ID)
	if _, err := engine.Check(session.ID, exercise.ID, "wrong"); err != nil {
		t.Fatal(err)
	}

	snapshot, err := engine.Progress(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Skills[exercise.Skill] = 999

	progress, err := engine.Progress(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := progress.Skills[exercise.Skill]; got != -1 {
		t.Fatalf("mutating progress snapshot changed engine skill score to %d", got)
	}
}

func TestWrongAnswerCostsHeartAndTargetsWeakSkill(t *testing.T) {
	engine := NewEngine()
	session := engine.NewSession()
	exercise, _ := engine.Next(session.ID)

	grade, err := engine.Check(session.ID, exercise.ID, "She walk to work every day.")
	if err != nil {
		t.Fatal(err)
	}
	if grade.Correct || grade.Hearts != 4 || grade.WeakestSkill != exercise.Skill {
		t.Fatalf("wrong answer was not reflected in grade: %+v", grade)
	}
	next, _ := engine.Next(session.ID)
	if next.Skill != exercise.Skill || next.ID == exercise.ID {
		t.Fatalf("expected a new targeted exercise for %s, got %+v", exercise.Skill, next)
	}
}

func TestNextRecyclesWeakSkillWhenItsUnseenExercisesAreExhausted(t *testing.T) {
	engine := NewEngine()
	engine.bank = []Exercise{
		{ID: "weak-1", Skill: "weak", answer: "one"},
		{ID: "weak-2", Skill: "weak", answer: "two"},
		{ID: "other-1", Skill: "other", answer: "other"},
	}
	session := engine.NewSession()

	first, _ := engine.Next(session.ID)
	if _, err := engine.Check(session.ID, first.ID, "wrong"); err != nil {
		t.Fatal(err)
	}
	second, _ := engine.Next(session.ID)
	if _, err := engine.Check(session.ID, second.ID, "wrong"); err != nil {
		t.Fatal(err)
	}

	next, err := engine.Next(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if next.Skill != "weak" {
		t.Fatalf("targeted practice switched to %q after weak-skill items were seen", next.Skill)
	}
	if next.ID == second.ID {
		t.Fatalf("targeted practice immediately repeated %q despite an alternative", next.ID)
	}
	if _, err := engine.Check(session.ID, next.ID, "wrong again"); err != nil {
		t.Fatalf("recycled weak-skill exercise %q was not gradeable: %v", next.ID, err)
	}
	after, err := engine.Next(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Skill != "weak" || after.ID == next.ID {
		t.Fatalf("repeated targeted practice did not rotate weak exercises: previous=%q next=%+v", next.ID, after)
	}
}

func TestNextRestartsCompletedWeakSkillBeforeSwitchingSkills(t *testing.T) {
	engine := NewEngine()
	engine.bank = []Exercise{
		{ID: "weak-1", Skill: "weak", answer: "right"},
		{ID: "other-1", Skill: "other", answer: "other"},
	}
	session := engine.NewSession()
	weak, _ := engine.Next(session.ID)
	if _, err := engine.Check(session.ID, weak.ID, "wrong"); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Check(session.ID, weak.ID, weak.answer); err != nil {
		t.Fatal(err)
	}

	next, err := engine.Next(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if next.Skill != "weak" {
		t.Fatalf("completed weak skill was abandoned for %q", next.Skill)
	}
	if _, err := engine.Check(session.ID, next.ID, next.answer); err != nil {
		t.Fatalf("restarted weak-skill exercise was not gradeable: %v", err)
	}
}

func TestNextStartsNewCycleAfterAllExercisesCompleted(t *testing.T) {
	engine := NewEngine()
	session := engine.NewSession()
	lastID := ""
	for range engine.bank {
		exercise, err := engine.Next(session.ID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := engine.Check(session.ID, exercise.ID, exercise.answer); err != nil {
			t.Fatal(err)
		}
		lastID = exercise.ID
	}

	exercise, err := engine.Next(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(engine.bank) > 1 && exercise.ID == lastID {
		t.Fatalf("new cycle immediately repeated %q", exercise.ID)
	}
	if _, err := engine.Check(session.ID, exercise.ID, exercise.answer); err != nil {
		t.Fatalf("new-cycle exercise %q was not gradeable: %v", exercise.ID, err)
	}
}

func TestNextReturnsErrorWhenExerciseBankIsEmpty(t *testing.T) {
	engine := NewEngine()
	engine.bank = nil
	session := engine.NewSession()

	if _, err := engine.Next(session.ID); err == nil {
		t.Fatal("expected an empty exercise bank error")
	}
}

func TestExerciseBankCoversCoreEnglishPractice(t *testing.T) {
	bank := defaultBank()
	if len(bank) < 12 {
		t.Fatalf("expected at least 12 exercises, got %d", len(bank))
	}
	kinds := map[string]bool{}
	skills := map[string]bool{}
	for _, exercise := range bank {
		kinds[exercise.Type] = true
		skills[exercise.Skill] = true
	}
	for _, kind := range []string{"multiple-choice", "sentence-build", "usage"} {
		if !kinds[kind] {
			t.Errorf("missing exercise type %q", kind)
		}
	}
	if len(skills) < 5 {
		t.Errorf("expected at least 5 skill areas, got %d", len(skills))
	}
}
