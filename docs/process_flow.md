# Programidex Initialization Process Flow

## Current Workflow

1. **Check for Existing Initialization**
    - If `.programidex/init_config.json` exists, load config.
    - If GitHub repo is missing, prompt to set up or enter manually.
    - If already initialized and complete, exit.

2. **Prompt for Project Type**
    - Ask if initializing an `app` or `module`.

3. **GitHub Setup**
    - Try to detect existing GitHub remote.
    - If not found, offer to set up GitHub repo (default: use current directory name) or enter manually.

4. **Go Module Setup**
    - Prompt for Go module path (e.g., `github.com/user/repo`).

5. **Blueprint Configuration**
    - For apps: Ask if it will have modules.
    - For modules: Ask for module name.
    - Always include Hugo site for now.
    - Always add `.programidex/` for config/logs.

6. **Display Proposed Config**
    - Show user the planned structure and settings.
    - Confirm before proceeding.

7. **Directory & Go Project Initialization**
    - Create all required directories.
    - Initialize `go.mod` if not present.

8. **Save Config & Log**
    - Save config as JSON in `.programidex/`.
    - Append actions to `.programidex/init.log`.

---

## Process Flow Diagram

```mermaid
flowchart TD
    A[Start] --> B{.programidex/init_config.json exists?}
    B -- Yes --> C{GitHub repo set?}
    C -- No --> D[Prompt to set up GitHub or enter manually]
    D --> E[Update config & log, exit]
    C -- Yes --> E
    B -- No --> F[Prompt for app or module]
    F --> G[GitHub setup (detect, create, or enter)]
    G --> H[Prompt for Go module path]
    H --> I[Configure blueprint (ask for modules or module name)]
    I --> J[Show config, confirm]
    J -- No --> K[Abort, log, exit]
    J -- Yes --> L[Create directories]
    L --> M[Initialize  if needed]
    M --> N[Save config & log]
    N --> O[Done]