(() => {
  "use strict";

  const requestedApi = new URLSearchParams(window.location.search).get("api");
  const localApiOrigins = ["http://localhost:8080", "http://127.0.0.1:8080"];
  const API_BASE = localApiOrigins.includes(requestedApi) ? requestedApi : "";
  const state = { session: null, exercise: null, selected: "", grade: null, loading: false };
  const $ = (id) => document.getElementById(id);
  const skillNames = {
    "subject-verb-agreement": "Grammar agreement",
    "past-tense": "Past tense",
    "context-vocabulary": "Vocabulary in context",
    "word-order": "Sentence building",
    articles: "Articles",
    prepositions: "Prepositions"
  };

  async function request(url, options = {}) {
    const response = await fetch(API_BASE + url, { headers: { "Content-Type": "application/json" }, ...options });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || "Something went wrong. Please try again.");
    return data;
  }

  async function start() {
    setBusy(true);
    try {
      state.session = await request("/api/sessions", { method: "POST" });
      updateStats(state.session);
      await loadNext();
    } catch (error) {
      showError(error.message);
    } finally {
      setBusy(false);
    }
  }

  async function loadNext() {
    setBusy(true);
    try {
      state.exercise = await request(`/api/sessions/${state.session.id}/exercise`);
      state.selected = "";
      state.grade = null;
      renderExercise();
    } catch (error) {
      showError(error.message);
    } finally {
      setBusy(false);
    }
  }

  function renderExercise() {
    const item = state.exercise;
    $("typeBadge").textContent = item.type.replaceAll("-", " ").toUpperCase();
    $("instruction").textContent = item.instruction;
    $("prompt").textContent = item.prompt;
    $("skillLabel").textContent = `Focus skill · ${item.skillLabel}`;
    $("choices").replaceChildren(...item.choices.map((choice, index) => {
      const button = document.createElement("button");
      button.className = "choice";
      button.type = "button";
      button.setAttribute("role", "radio");
      button.setAttribute("aria-checked", "false");
      const key = document.createElement("span");
      key.className = "choice-key";
      key.textContent = String(index + 1);
      const text = document.createElement("span");
      text.textContent = choice;
      button.append(key, text);
      button.addEventListener("click", () => selectAnswer(choice, button));
      return button;
    }));
    $("feedback").hidden = true;
    $("retryButton").hidden = true;
    $("nextButton").hidden = true;
    $("checkButton").hidden = false;
    $("checkButton").disabled = true;
    $("formError").textContent = "";
  }

  function selectAnswer(choice, button) {
    if (state.grade || state.loading) return;
    state.selected = choice;
    document.querySelectorAll(".choice").forEach((element) => {
      const active = element === button;
      element.classList.toggle("selected", active);
      element.setAttribute("aria-checked", String(active));
    });
    $("checkButton").disabled = false;
    $("formError").textContent = "";
  }

  async function checkAnswer() {
    if (!state.selected) {
      $("formError").textContent = "Choose an answer before checking.";
      return;
    }
    setBusy(true);
    try {
      state.grade = await request(`/api/sessions/${state.session.id}/answers`, {
        method: "POST",
        body: JSON.stringify({ exerciseId: state.exercise.id, answer: state.selected })
      });
      state.session.xp = state.grade.totalXp;
      state.session.hearts = state.grade.hearts;
      state.session.answered += 1;
      if (state.grade.correct) state.session.correct += 1;
      updateStats(state.session);
      showFeedback();
      await refreshProgress();
    } catch (error) {
      $("formError").textContent = error.message;
    } finally {
      setBusy(false);
    }
  }

  function showFeedback() {
    const grade = state.grade;
    const panel = $("feedback");
    panel.hidden = false;
    panel.className = `feedback ${grade.correct ? "correct" : "incorrect"}`;
    $("feedbackIcon").textContent = grade.correct ? "✓" : "×";
    $("feedbackTitle").textContent = grade.correct ? `Correct! +${grade.xpEarned} XP` : "Good try — here’s the pattern";
    $("feedbackCopy").textContent = grade.explanation;
    $("answerReveal").textContent = grade.correct ? "" : `Correct answer: ${grade.correctAnswer}`;
    $("checkButton").hidden = true;
    $("nextButton").hidden = false;
    $("retryButton").hidden = grade.correct;
    document.querySelectorAll(".choice").forEach((button) => { button.disabled = true; });
  }

  function retry() {
    state.grade = null;
    state.selected = "";
    document.querySelectorAll(".choice").forEach((button) => {
      button.disabled = false;
      button.classList.remove("selected");
      button.setAttribute("aria-checked", "false");
    });
    $("feedback").hidden = true;
    $("retryButton").hidden = true;
    $("nextButton").hidden = true;
    $("checkButton").hidden = false;
    $("checkButton").disabled = true;
  }

  function updateStats(progress) {
    $("xp").textContent = progress.xp;
    $("dailyXp").textContent = Math.min(progress.xp, 50);
    $("hearts").textContent = progress.hearts;
    const completed = Math.min(progress.answered, 10);
    $("lessonProgress").style.width = `${completed * 10}%`;
    $("progressText").textContent = `${completed} / 10`;
  }

  async function refreshProgress() {
    const progress = await request(`/api/sessions/${state.session.id}/progress`);
    state.session = progress;
    updateStats(progress);
    renderDashboard(progress);
  }

  function renderDashboard(progress) {
    $("accuracy").textContent = progress.answered ? `${Math.round(progress.correct / progress.answered * 100)}%` : "—";
    const entries = Object.entries(skillNames).map(([key, name]) => [key, name, progress.skills[key] || 0]);
    $("skillGrid").replaceChildren(...entries.map(([key, name, score]) => {
      const card = document.createElement("article");
      card.className = "skill-card";
      const confidence = Math.max(10, Math.min(100, 50 + score * 20));
      card.innerHTML = `<header><strong>${name}</strong><span>${score > 0 ? "Growing" : score < 0 ? "Needs focus" : "Not sampled"}</span></header><p>${score < 0 ? "A targeted follow-up is queued." : "Practice score: " + score}</p><div class="skill-meter"><span style="width:${confidence}%"></span></div>`;
      card.dataset.skill = key;
      return card;
    }));
    const weakest = entries.sort((a, b) => a[2] - b[2])[0];
    if (progress.answered) {
      $("focusTitle").textContent = weakest[2] < 0 ? `Next focus: ${weakest[1]}` : "Balanced start — keep sampling skills";
      $("focusCopy").textContent = weakest[2] < 0 ? "Your next unseen item stays in this skill so you can apply the feedback immediately." : "As answers arrive, the lowest-scoring skill becomes the next practice target.";
    }
  }

  function setBusy(value) {
    state.loading = value;
    $("checkButton").disabled = value || !state.selected;
    $("nextButton").disabled = value;
  }

  function showError(message) {
    $("prompt").textContent = "Practice paused";
    $("skillLabel").textContent = message;
    $("choices").innerHTML = '<div class="error-state">Make sure the Go API is running, then refresh the page.</div>';
  }

  function showView(name) {
    const learn = name === "learn";
    $("learnView").hidden = !learn;
    $("progressView").hidden = learn;
    $("actionBar").hidden = !learn;
    document.querySelectorAll(".nav-item").forEach((item) => item.classList.toggle("active", item.dataset.view === name));
    setNavOpen(false);
    if (!learn && state.session) refreshProgress().catch((error) => showError(error.message));
  }

  // Below 800px the sidebar is hidden, so the menu button must reveal it as an overlay
  // instead of navigating away — otherwise the learner cannot get back to the lesson.
  function setNavOpen(open) {
    $("app").classList.toggle("nav-open", open);
    $("menuButton").setAttribute("aria-expanded", String(open));
    $("menuButton").setAttribute("aria-label", open ? "Close menu" : "Open menu");
  }

  $("checkButton").addEventListener("click", checkAnswer);
  $("nextButton").addEventListener("click", loadNext);
  $("retryButton").addEventListener("click", retry);
  document.querySelectorAll(".nav-item").forEach((item) => item.addEventListener("click", () => showView(item.dataset.view)));
  $("menuButton").addEventListener("click", () => setNavOpen(!$("app").classList.contains("nav-open")));
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") setNavOpen(false);
    if (event.key >= "1" && event.key <= "9" && !state.grade) {
      const button = document.querySelectorAll(".choice")[Number(event.key) - 1];
      if (button) button.click();
    }
    if (event.key === "Enter" && !$("checkButton").hidden && !$("checkButton").disabled) checkAnswer();
  });

  renderDashboard({ answered: 0, correct: 0, skills: {} });
  start();
})();
