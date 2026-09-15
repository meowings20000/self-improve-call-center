# AI Technical Note — LingoLift

## Product and learning loop

LingoLift is deliberately **AI-shaped, not AI-dependent**: it demonstrates an adaptive tutoring loop using deterministic, inspectable rules rather than a generative model. A learner receives one authored English prompt, chooses an answer, receives immediate task-level feedback, and can retry or continue. The backend updates a per-skill signal after every valid attempt. A wrong answer queues a new unseen item from that same lowest-scoring skill, so the follow-up asks the learner to apply the explanation rather than merely reread it.

The design maps directly to **practice testing** in Dunlosky et al. (2013): every screen requires retrieval or discrimination before revealing the answer. Their review assessed practice testing and distributed practice as high-utility techniques across a broad range of learning conditions, learners, materials, and outcome tasks. This PoC implements the practice-testing portion; it does not claim to implement long-term distributed scheduling. [Dunlosky et al., 2013, DOI 10.1177/1529100612453266](https://doi.org/10.1177/1529100612453266).

Feedback maps to Hattie and Timperley’s (2007) model: after a response, LingoLift reports the result (where the learner is), explains the applicable language rule (how to close the task gap), and identifies the targeted next focus (where to go next). Their review stresses that feedback’s effect varies with its type and delivery, so the interface avoids generic praise alone and keeps the explanation tied to the task. [Hattie & Timperley, 2007, DOI 10.3102/003465430298487](https://doi.org/10.3102/003465430298487).

## Architecture and data flow

The browser client is plain HTML, CSS, and JavaScript with no remote assets. Nginx serves it and proxies `/api` to a Go service. The Go standard-library HTTP layer owns request validation and status codes; the `learning` package owns the exercise bank, session state, grading, XP/hearts, and selection rule. Answers and explanations are unexported Go fields, so the exercise JSON cannot expose them. They appear only in the grading response after a learner attempt.

`POST /api/sessions` creates an in-memory session. `GET .../exercise` selects a question. `POST .../answers` validates non-empty input and grades exact authored choices. Correct answers add 10 XP and raise the skill signal; incorrect answers cost one heart (floored at zero) and lower it. `GET .../progress` provides only the values needed for the explanatory dashboard. No probabilistic service, API key, database, or external network is required.

## Why deterministic adaptation

For a small assessed interaction, deterministic behavior is easier to test, explain, reproduce, and audit than an LLM-generated exercise stream. The adaptation is intentionally modest: prioritize an unseen exercise in the weakest sampled skill, otherwise continue through the authored bank. This is a formative routing heuristic, not a learner model. Skill values explain the recommendation but are not calibrated ability estimates.

A production evolution could use an expert-reviewed item model, response-time and difficulty metadata, spaced review, and a transparent mastery estimator. Generative AI could draft candidate distractors or feedback offline, but a human reviewer should approve them before learners see them. Sending free-text learner responses to a model would require informed consent, retention controls, vendor review, filtering, and a non-AI fallback.

## How AI tools were used

- **Model-drafted work:** the coding model drafted the Go exercise engine and API, the responsive browser UI, initial authored question bank, automated tests, and documentation; every executable path was then run locally and reviewed.
- **Suggestion rejected:** an unbounded LLM tutor and live AI-generated questions were rejected because they would require keys, make grading less reproducible, and could expose learner input. The assessed interaction instead uses inspectable authored answers and deterministic routing.
- **Human/agent changes:** the design was narrowed to one complete learning loop, grading answers were kept out of exercise responses, input limits and error states were added, research claims were checked against their DOI records, and the full wrong-answer → feedback → targeted-follow-up route was browser-tested.
- **Next step:** add expert-reviewed difficulty metadata and spaced review, then validate whether targeted follow-ups improve delayed performance with learners—not merely same-session accuracy.

## Safety, privacy, and limitations

The PoC collects no identity, uses no analytics, and keeps session attempts only in process memory. It should not receive personal or sensitive text. Restarting the API removes all state, and the in-memory store evicts the oldest session after 1,000 entries. Session IDs are cryptographically random; direct browser access to the API accepts only the documented local-development origin, while Docker exposes both services on loopback only. The fixed bank is small, English-only, and not psychometrically validated. XP, hearts, accuracy, and skill bars are motivational/formative UI—not grades, certification, CEFR placement, or evidence for employment or education decisions. Before real learner use, the product needs content-specialist review, assistive-technology testing, broader item coverage, abuse testing, age-appropriate privacy review, and user-controlled data export/deletion if persistence is added.

## References

- Dunlosky, J., Rawson, K. A., Marsh, E. J., Nathan, M. J., & Willingham, D. T. (2013). *Improving Students’ Learning With Effective Learning Techniques: Promising Directions From Cognitive and Educational Psychology*. Psychological Science in the Public Interest, 14(1), 4–58. https://doi.org/10.1177/1529100612453266
- Hattie, J., & Timperley, H. (2007). *The Power of Feedback*. Review of Educational Research, 77(1), 81–112. https://doi.org/10.3102/003465430298487
