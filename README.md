# gitext

A safe git workflow automation CLI tool for engineering teams. `gitext` replaces manual git workflow steps with safe, repeatable commands that enforce branch protection rules and prevent accidental production contamination.

## Features

- **Safe branch management**: Enforces rules to prevent accidental production contamination from stage
- **Automated workflows**: Common git operations simplified into single commands
- **Branch protection**: Pre-push hooks prevent direct pushes to protected branches
- **CI integration**: Run configured CI checks before creating PRs
- **Smart suggestions**: Commands suggest next steps based on current state
- **AI-powered commit messages**: Generate commit messages automatically using AI (OpenAI or OpenRouter) following Conventional Commits specification

## Installation

### Install Latest Release

**Linux (amd64):**
```bash
curl -L https://github.com/imemir/gitext/releases/latest/download/gitext-linux-amd64 -o gitext
chmod +x gitext
sudo mv gitext /usr/local/bin/
```

**Linux (arm64):**
```bash
curl -L https://github.com/imemir/gitext/releases/latest/download/gitext-linux-arm64 -o gitext
chmod +x gitext
sudo mv gitext /usr/local/bin/
```

**macOS (amd64):**
```bash
curl -L https://github.com/imemir/gitext/releases/latest/download/gitext-darwin-amd64 -o gitext
chmod +x gitext
sudo mv gitext /usr/local/bin/
```

**macOS (arm64 / Apple Silicon):**
```bash
curl -L https://github.com/imemir/gitext/releases/latest/download/gitext-darwin-arm64 -o gitext
chmod +x gitext
sudo mv gitext /usr/local/bin/
```

**Windows (amd64):**
```powershell
Invoke-WebRequest -Uri https://github.com/imemir/gitext/releases/latest/download/gitext-windows-amd64.exe -OutFile gitext.exe
# Add to PATH or move to desired location
```

**Windows (arm64):**
```powershell
Invoke-WebRequest -Uri https://github.com/imemir/gitext/releases/latest/download/gitext-windows-arm64.exe -OutFile gitext.exe
# Add to PATH or move to desired location
```

### Build from source

```bash
git clone https://github.com/imemir/gitext.git
cd gitext
go build -o gitext ./cmd/gitext
sudo mv gitext /usr/local/bin/
```

### Verify installation

```bash
gitext --help
```

### Shell Completion

gitext supports shell completion for bash, zsh, fish, and PowerShell.

**Bash:**
```bash
source <(gitext completion bash)
# Add to ~/.bashrc or ~/.bash_profile for persistence
```

**Zsh:**
```bash
source <(gitext completion zsh)
# Add to ~/.zshrc for persistence
```

**Fish:**
```bash
gitext completion fish | source
# Add to ~/.config/fish/config.fish for persistence
```

**PowerShell:**
```powershell
gitext completion powershell | Out-String | Invoke-Expression
# Add to your PowerShell profile for persistence
```

## Quick Start

1. **Initialize gitext in your repository:**

```bash
gitext init --install-hooks
```

This creates a `.gitext` configuration file and optionally installs git hooks.

2. **Check your current status:**

```bash
gitext status
```

3. **Start a new feature branch:**

```bash
gitext start feature --ticket KWS-123 --slug retry-policy --from stage
```

4. **Optional: Setup AI for commit messages:**

```bash
gitext ai setup
```

After setup, you can use `gitext commit` instead of `git commit` to automatically generate commit messages.

## Configuration

The `.gitext` file is a YAML configuration file placed in your repository root. Here's an example:

```yaml
branch:
  production: "production"
  stage: "stage"
naming:
  feature: "feature/*"
  hotfix: "hotfix/*"
merge:
  requireRetargetForProdFromStage: true
ci:
  stage:
    - "go test ./..."
    - "go vet ./..."
  production:
    - "go test ./..."
    - "go vet ./..."
    - "go build ./..."
pr:
  templatePath: ".github/pull_request_template.md"  # optional
remote:
  name: "origin"
```

### Configuration Fields

- **branch.production**: Name of the production branch (default: "production")
- **branch.stage**: Name of the stage branch (default: "stage")
- **naming.feature**: Pattern for feature branch names (default: "feature/*")
- **naming.hotfix**: Pattern for hotfix branch names (default: "hotfix/*")
- **merge.requireRetargetForProdFromStage**: Enforce retargeting workflow (default: true)
- **ci.stage**: Array of shell commands to run before PRs to stage
- **ci.production**: Array of shell commands to run before PRs to production
- **pr.templatePath**: Optional path to PR template file (relative to repo root)
- **remote.name**: Git remote name (default: "origin")

### AI Configuration

AI configuration is stored in `~/.gitext/config.yaml` (global configuration, not per-repository). This file is created automatically when you run `gitext ai setup`.

**Example configuration:**

```yaml
provider: openai  # or "openrouter"
openai:
  api_key: sk-...
  model: gpt-4o
openrouter:
  api_key: sk-...
  model: google/gemini-flash-1.5-8b
  use_free_model: true
```

**Security:**
- The config file is created with permissions `0600` (read/write for owner only)
- API keys are masked when displayed (`gitext ai config`)
- You can reconfigure anytime with `gitext ai setup`

## Commands

### `gitext init`

Initialize gitext configuration in the current repository.

```bash
gitext init [--install-hooks]
```

- Creates `.gitext` configuration file if it doesn't exist
- `--install-hooks`: Install pre-push git hooks to prevent direct pushes to protected branches

### `gitext status`

Show current git status and suggest next steps.

```bash
gitext status
```

Displays:
- Current branch
- Working tree state (clean/dirty)
- Ahead/behind status vs remote, stage, and production
- Suggested next command

### `gitext sync <target>`

Safely sync a branch with its remote using fast-forward only.

```bash
gitext sync stage
gitext sync production
```

- Fetches from remote
- Pulls with `--ff-only` (fails if fast-forward not possible)
- Suggests update command if branch has diverged

### `gitext start feature`

Create a new feature branch from stage or production.

```bash
gitext start feature --ticket KWS-123 --slug retry-policy --from stage
```

- Validates branch name matches configured pattern
- Creates branch from specified source (stage or production)
- Checks out the new branch

### `gitext update feature`

Update current feature branch with changes from stage or production.

```bash
gitext update feature --with stage --mode rebase
gitext update feature --with production --mode merge
```

- `--with`: Source branch (stage or production)
- `--mode`: Update method (rebase or merge, default: rebase)

### `gitext retarget feature`

Retarget a feature branch from stage onto production.

```bash
gitext retarget feature --onto production --from stage
```

**Safety features:**
- Validates current branch is a feature branch (unless `--override`)
- Detects shared branches (multiple authors) and requires `--i-know-what-im-doing`
- Uses `git rebase --onto` to rewrite history
- Warns about force push requirements

**Flags:**
- `--override`: Allow retargeting non-feature branches
- `--i-know-what-im-doing`: Bypass shared branch safety check

### `gitext prepare pr`

Run CI checks and generate PR text.

```bash
gitext prepare pr --to stage
gitext prepare pr --to production
```

- Runs configured CI commands for the target branch
- Generates PR text with branch info, ticket, and commit summary
- Prints PR text to stdout
- Uses template if configured

### `gitext cleanup`

Clean up merged local branches.

```bash
gitext cleanup [--hard]
```

- Lists branches merged into stage or production
- Dry-run by default (shows what would be deleted)
- `--hard`: Actually delete branches
- Skips protected branches (stage/production)

### `gitext completion`

Generate shell completion scripts for bash, zsh, fish, or PowerShell.

```bash
gitext completion bash
gitext completion zsh
gitext completion fish
gitext completion powershell
```

See the [Shell Completion](#shell-completion) section for installation instructions.

## AI Commands

gitext includes AI-powered features for generating commit messages automatically. The AI analyzes your code changes and generates commit messages following the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.

### Setup AI Provider

Before using AI features, you need to configure an AI provider:

```bash
gitext ai setup
```

This interactive command will:
1. Let you choose between OpenAI or OpenRouter
2. Prompt for your API key (input is hidden)
3. Let you select a model
4. Test the connection
5. Save configuration to `~/.gitext/config.yaml`

**OpenAI Options:**
- `gpt-4o` (default - recommended)
- `gpt-4o-mini` (faster, cheaper)
- `gpt-4-turbo`
- `gpt-3.5-turbo` (cheapest)
- Custom model

**OpenRouter Options:**
- Free models: `google/gemini-flash-1.5-8b`, `qwen/qwen-2.5-7b-instruct`, `mistralai/mistral-7b-instruct-v0.2`
- Custom model (any model supported by OpenRouter)

### `gitext ai config`

View your current AI configuration or test the connection.

```bash
# View configuration
gitext ai config

# Test connection
gitext ai config --test
```

Displays:
- Current provider (OpenAI or OpenRouter)
- Masked API key (for security)
- Selected model
- Configuration file path

### `gitext commit`

Generate a commit message using AI and create the commit.

```bash
# Stage your changes first
git add .

# Generate commit message with AI and commit
gitext commit

# Use a custom message instead
gitext commit --message "fix: custom commit message"
```

**How it works:**
1. Checks for staged changes
2. Gets the diff of staged changes
3. Sends diff to AI provider
4. Generates commit message following Conventional Commits format
5. Shows the generated message and asks for confirmation
6. Creates the commit if confirmed

**Example output:**
```
→ Getting staged changes
→ Generating commit message with AI...
✓ Generated commit message:

  feat(auth): add password reset functionality

Create commit with this message? [Y/n]: y
→ Creating commit
✓ Commit created successfully
```

**Note:** The AI analyzes your code changes and generates messages in the format `type(scope): description` where:
- `type`: feat, fix, docs, style, refactor, perf, test, chore, etc.
- `scope`: optional, the area affected (e.g., auth, api, ui)
- `description`: brief summary in imperative mood

## Example Workflows

### Staging-First Workflow

1. **Start a feature from stage:**

```bash
gitext start feature --ticket KWS-123 --slug new-feature --from stage
```

2. **Make changes and commit with AI:**

```bash
git add .
gitext commit
```

Or use traditional git commit:
```bash
git add .
git commit -m "Add new feature"
```

3. **Update with latest stage changes:**

```bash
gitext update feature --with stage --mode rebase
```

4. **Prepare PR to stage:**

```bash
gitext prepare pr --to stage
```

5. **After PR is merged to stage, retarget for production:**

```bash
gitext retarget feature --onto production --from stage
```

6. **Prepare PR to production:**

```bash
gitext prepare pr --to production
```

### Hotfix from Production

1. **Start hotfix from production:**

```bash
gitext start feature --ticket HOTFIX-456 --slug critical-fix --from production
```

2. **Make fix and commit with AI:**

```bash
git add .
gitext commit
```

Or use traditional git commit:
```bash
git add .
git commit -m "Fix critical issue"
```

3. **Prepare PR to production:**

```bash
gitext prepare pr --to production
```

4. **After merge, sync production to stage:**

```bash
gitext sync production
gitext sync stage
```

## Safety Features

1. **No destructive operations without flags**: Commands require explicit flags (`--hard`, `--force`, `--i-know-what-im-doing`) for destructive operations
2. **Pre-push hooks**: Blocks direct pushes to protected branches (unless CI user detected)
3. **Working tree checks**: Most commands fail if working tree is dirty
4. **Fast-forward only**: Default to safe merge strategies (`--ff-only`)
5. **Shared branch detection**: Warns/blocks retargeting shared branches
6. **Dry-run mode**: Global `--dry-run` flag shows what would be done without executing

## Global Flags

- `--dry-run`: Show what would be done without executing
- `--verbose`: Show detailed git command output

## Troubleshooting

### "Working tree dirty"

Commit or stash your changes first:

```bash
git commit -am "Your message"
# or
git stash
```

### "Fast-forward not possible"

Your branch has diverged. Update it first:

```bash
gitext update feature --with stage
```

### "Branch appears shared"

The branch has multiple authors and retargeting would rewrite shared history. Use `--i-know-what-im-doing` if you're certain:

```bash
gitext retarget feature --onto production --from stage --i-know-what-im-doing
```

### "Remote not configured"

Add the remote:

```bash
git remote add origin <repository-url>
```

### "AI configuration not found"

Set up AI provider first:

```bash
gitext ai setup
```

### "Failed to generate commit message"

Possible causes:
- Invalid API key: Run `gitext ai config --test` to verify
- Network issues: Check your internet connection
- API rate limits: Wait a moment and try again
- No staged changes: Stage your changes with `git add` first

### "Connection test failed"

- Verify your API key is correct
- Check if you have sufficient API credits/quota
- For OpenRouter: Ensure the model name is correct
- Try running `gitext ai setup` again to reconfigure

## Contributing

Contributions are welcome! Please ensure all tests pass and follow the existing code style.

## License

MIT License - see [LICENSE](LICENSE) file for details.

