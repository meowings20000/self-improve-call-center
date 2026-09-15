package learning

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
)

type Exercise struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Skill       string   `json:"skill"`
	SkillLabel  string   `json:"skillLabel"`
	Prompt      string   `json:"prompt"`
	Instruction string   `json:"instruction"`
	Choices     []string `json:"choices"`
	answer      string
	explanation string
}

type Session struct {
	ID         string         `json:"id"`
	Hearts     int            `json:"hearts"`
	XP         int            `json:"xp"`
	Answered   int            `json:"answered"`
	Correct    int            `json:"correct"`
	Skills     map[string]int `json:"skills"`
	seen       map[string]bool
	completed  map[string]bool
	lastIssued string
}

type Grade struct {
	Correct       bool   `json:"correct"`
	Feedback      string `json:"feedback"`
	Explanation   string `json:"explanation"`
	CorrectAnswer string `json:"correctAnswer"`
	XPEarned      int    `json:"xpEarned"`
	TotalXP       int    `json:"totalXp"`
	Hearts        int    `json:"hearts"`
	WeakestSkill  string `json:"weakestSkill"`
}

type Engine struct {
	mu           sync.Mutex
	sessions     map[string]*Session
	sessionOrder []string
	maxSessions  int
	bank         []Exercise
}

func NewEngine() *Engine {
	return newEngine(1000)
}

func newEngine(maxSessions int) *Engine {
	return &Engine{sessions: make(map[string]*Session), maxSessions: maxSessions, bank: defaultBank()}
}

func (e *Engine) NewSession() Session {
	e.mu.Lock()
	defer e.mu.Unlock()
	var id string
	for id == "" || e.sessions[id] != nil {
		token := make([]byte, 16)
		if _, err := rand.Read(token); err != nil {
			panic("secure session ID generation failed: " + err.Error())
		}
		id = hex.EncodeToString(token)
	}
	s := &Session{ID: id, Hearts: 5, Skills: map[string]int{}, seen: map[string]bool{}, completed: map[string]bool{}}
	if len(e.sessions) >= e.maxSessions {
		delete(e.sessions, e.sessionOrder[0])
		e.sessionOrder = e.sessionOrder[1:]
	}
	e.sessions[s.ID] = s
	e.sessionOrder = append(e.sessionOrder, s.ID)
	snapshot := *s
	snapshot.Skills = make(map[string]int)
	return snapshot
}

func (e *Engine) Next(sessionID string) (Exercise, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	s, ok := e.sessions[sessionID]
	if !ok {
		return Exercise{}, errors.New("session not found")
	}
	if len(e.bank) == 0 {
		return Exercise{}, errors.New("no exercises available")
	}
	issue := func(item Exercise) (Exercise, error) {
		s.seen[item.ID] = true
		s.lastIssued = item.ID
		return item, nil
	}
	weakest := ""
	weakestScore := 1
	for skill, score := range s.Skills {
		if score < weakestScore {
			weakest, weakestScore = skill, score
		}
	}
	if weakest != "" {
		for _, item := range e.bank {
			if item.Skill == weakest && !s.seen[item.ID] {
				return issue(item)
			}
		}
		for _, item := range e.bank {
			if item.Skill == weakest && !s.completed[item.ID] && item.ID != s.lastIssued {
				return issue(item)
			}
		}
		for _, item := range e.bank {
			if item.Skill == weakest && !s.completed[item.ID] {
				return issue(item)
			}
		}
		for _, item := range e.bank {
			if item.Skill == weakest {
				delete(s.seen, item.ID)
				delete(s.completed, item.ID)
			}
		}
		for _, item := range e.bank {
			if item.Skill == weakest && item.ID != s.lastIssued {
				return issue(item)
			}
		}
		for _, item := range e.bank {
			if item.Skill == weakest {
				return issue(item)
			}
		}
	}
	for _, item := range e.bank {
		if !s.seen[item.ID] {
			return issue(item)
		}
	}
	for _, item := range e.bank {
		if !s.completed[item.ID] && item.ID != s.lastIssued {
			return issue(item)
		}
	}
	for _, item := range e.bank {
		if !s.completed[item.ID] {
			return issue(item)
		}
	}
	s.seen = make(map[string]bool)
	s.completed = make(map[string]bool)
	for _, item := range e.bank {
		if item.ID != s.lastIssued {
			return issue(item)
		}
	}
	return issue(e.bank[0])
}

func (e *Engine) Check(sessionID, exerciseID, answer string) (Grade, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	s, ok := e.sessions[sessionID]
	if !ok {
		return Grade{}, errors.New("session not found")
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return Grade{}, errors.New("answer cannot be empty")
	}
	var exercise *Exercise
	for i := range e.bank {
		if e.bank[i].ID == exerciseID {
			exercise = &e.bank[i]
			break
		}
	}
	if exercise == nil {
		return Grade{}, errors.New("exercise not found")
	}
	if !s.seen[exerciseID] {
		return Grade{}, errors.New("exercise not issued to session")
	}
	if s.completed[exerciseID] {
		return Grade{}, errors.New("exercise already completed")
	}
	correct := answer == exercise.answer
	earned := 0
	feedback := "Not quite — review the rule and try a targeted question."
	weakestSkill := ""
	if correct {
		s.completed[exerciseID] = true
		earned = 10
		feedback = "Excellent! You applied the rule correctly."
		s.Correct++
		s.Skills[exercise.Skill]++
	} else {
		s.Hearts--
		if s.Hearts < 0 {
			s.Hearts = 0
		}
		s.Skills[exercise.Skill]--
		weakestSkill = exercise.Skill
	}
	s.Answered++
	s.XP += earned
	return Grade{Correct: correct, Feedback: feedback, Explanation: exercise.explanation, CorrectAnswer: exercise.answer, XPEarned: earned, TotalXP: s.XP, Hearts: s.Hearts, WeakestSkill: weakestSkill}, nil
}

func (e *Engine) Progress(sessionID string) (Session, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	s, ok := e.sessions[sessionID]
	if !ok {
		return Session{}, errors.New("session not found")
	}
	snapshot := *s
	snapshot.Skills = make(map[string]int, len(s.Skills))
	for skill, score := range s.Skills {
		snapshot.Skills[skill] = score
	}
	return snapshot, nil
}

func defaultBank() []Exercise {
	return []Exercise{
		{ID: "grammar-1", Type: "multiple-choice", Skill: "subject-verb-agreement", SkillLabel: "Grammar agreement", Instruction: "Choose the correct sentence", Prompt: "Which sentence is correct?", Choices: []string{"She walk to work every day.", "She walks to work every day.", "She walking to work every day."}, answer: "She walks to work every day.", explanation: "With he, she, or it in the simple present, add -s to the verb."},
		{ID: "grammar-2", Type: "multiple-choice", Skill: "subject-verb-agreement", SkillLabel: "Grammar agreement", Instruction: "Choose the verb", Prompt: "My brother ___ coffee every morning.", Choices: []string{"drink", "drinks", "drinking"}, answer: "drinks", explanation: "My brother is third-person singular, so the simple-present verb takes -s."},
		{ID: "grammar-3", Type: "multiple-choice", Skill: "subject-verb-agreement", SkillLabel: "Grammar agreement", Instruction: "Complete the sentence", Prompt: "The dogs ___ loudly at night.", Choices: []string{"barks", "bark", "is barking"}, answer: "bark", explanation: "A plural subject takes the base verb in the simple present."},
		{ID: "tense-1", Type: "multiple-choice", Skill: "past-tense", SkillLabel: "Past tense", Instruction: "Choose the correct past form", Prompt: "Yesterday, we ___ a new café.", Choices: []string{"visit", "visited", "visiting"}, answer: "visited", explanation: "Yesterday signals the simple past; regular verbs add -ed."},
		{ID: "tense-2", Type: "multiple-choice", Skill: "past-tense", SkillLabel: "Past tense", Instruction: "Complete the sentence", Prompt: "Last night, Maya ___ an interesting book.", Choices: []string{"read", "reads", "reading"}, answer: "read", explanation: "The past form of read is spelled read (and pronounced 'red')."},
		{ID: "tense-3", Type: "usage", Skill: "past-tense", SkillLabel: "Past tense", Instruction: "Choose the natural sentence", Prompt: "Which sentence describes a finished trip?", Choices: []string{"I go to Paris last year.", "I went to Paris last year.", "I am go to Paris last year."}, answer: "I went to Paris last year.", explanation: "Went is the irregular past form of go."},
		{ID: "vocab-1", Type: "multiple-choice", Skill: "context-vocabulary", SkillLabel: "Vocabulary in context", Instruction: "Choose the closest meaning", Prompt: "The instructions were concise, so everyone understood them quickly. Concise means…", Choices: []string{"brief and clear", "confusing", "very loud"}, answer: "brief and clear", explanation: "Concise language communicates an idea clearly in few words."},
		{ID: "vocab-2", Type: "usage", Skill: "context-vocabulary", SkillLabel: "Vocabulary in context", Instruction: "Choose the best word", Prompt: "The team was ___ after finishing the difficult project.", Choices: []string{"exhausted", "ancient", "careless"}, answer: "exhausted", explanation: "Exhausted means extremely tired, which fits after difficult work."},
		{ID: "vocab-3", Type: "usage", Skill: "context-vocabulary", SkillLabel: "Vocabulary in context", Instruction: "Choose the natural use", Prompt: "Which sentence uses 'reliable' correctly?", Choices: []string{"This clock is reliable; it always shows the right time.", "The soup tastes reliable.", "She reliable ran home."}, answer: "This clock is reliable; it always shows the right time.", explanation: "Reliable describes someone or something consistently dependable."},
		{ID: "build-1", Type: "sentence-build", Skill: "word-order", SkillLabel: "Sentence building", Instruction: "Build the sentence in the correct order", Prompt: "often / on Sundays / we / hike", Choices: []string{"Often we hike on Sundays.", "We often hike on Sundays.", "We hike on often Sundays."}, answer: "We often hike on Sundays.", explanation: "Frequency adverbs usually go before the main verb: we often hike."},
		{ID: "build-2", Type: "sentence-build", Skill: "word-order", SkillLabel: "Sentence building", Instruction: "Build the question", Prompt: "you / where / live / do", Choices: []string{"Where do you live?", "Where you do live?", "Do where you live?"}, answer: "Where do you live?", explanation: "Wh- questions use question word + auxiliary + subject + base verb."},
		{ID: "build-3", Type: "sentence-build", Skill: "word-order", SkillLabel: "Sentence building", Instruction: "Build the sentence", Prompt: "has / already / finished / Lena", Choices: []string{"Already Lena has finished.", "Lena has already finished.", "Lena already has finish."}, answer: "Lena has already finished.", explanation: "Already normally follows the auxiliary has in the present perfect."},
		{ID: "usage-1", Type: "usage", Skill: "articles", SkillLabel: "Articles", Instruction: "Choose the correct article", Prompt: "I saw ___ elephant at the zoo.", Choices: []string{"a", "an", "the"}, answer: "an", explanation: "Use an before a vowel sound, as in elephant."},
		{ID: "usage-2", Type: "multiple-choice", Skill: "articles", SkillLabel: "Articles", Instruction: "Complete the sentence", Prompt: "Could you close ___ window next to you?", Choices: []string{"a", "an", "the"}, answer: "the", explanation: "Use the for a specific window identified by 'next to you'."},
		{ID: "usage-3", Type: "usage", Skill: "prepositions", SkillLabel: "Prepositions", Instruction: "Choose the correct preposition", Prompt: "The meeting starts ___ 9:00 a.m.", Choices: []string{"in", "on", "at"}, answer: "at", explanation: "Use at with a precise clock time."},
		{ID: "usage-4", Type: "usage", Skill: "prepositions", SkillLabel: "Prepositions", Instruction: "Choose the correct preposition", Prompt: "We have class ___ Monday.", Choices: []string{"at", "on", "in"}, answer: "on", explanation: "Use on with days of the week."},
	}
}
