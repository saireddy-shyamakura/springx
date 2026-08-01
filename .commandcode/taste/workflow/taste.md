# Workflow Preferences

- Wants a thorough security audit (covering all types of bugs/vulnerabilities) performed on the repo before anything is pushed to production. Confidence: 0.8
- Prefers the agent to work autonomously through long multi-step tasks and continue without pausing for confirmation at each step (replied "continue" repeatedly while the agent explored and audited). Confidence: 0.6
- When an audit/review produces findings, wants ALL of them fixed — complete remediation of the full findings list, including low-severity and hygiene issues (e.g. pinned CI tool versions, hardcoded passwords, doc checksums) — not just the top-priority subset the agent suggests starting with. Confidence: 0.7
