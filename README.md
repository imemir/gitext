# gitext

A safe git workflow automation CLI tool for engineering teams. `gitext` replaces manual git workflow steps with safe, repeatable commands that enforce branch protection rules and prevent accidental production contamination.

## Features

- **Safe branch management**: Enforces rules to prevent accidental production contamination from stage
- **Automated workflows**: Common git operations simplified into single commands
- **Branch protection**: Pre-push hooks prevent direct pushes to protected branches
- **CI integration**: Run configured CI checks before creating PRs
- **Smart suggestions**: Commands suggest next steps based on current state

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

## Example Workflows

### Staging-First Workflow

1. **Start a feature from stage:**

```bash
gitext start feature --ticket KWS-123 --slug new-feature --from stage
```

2. **Make changes and commit:**

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

2. **Make fix and commit:**

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

## Releasing

To create a new release:

1. Make sure you're on the `master` branch with a clean working tree
2. Run `make release` to:
   - Analyze changes since the last release
   - Determine the next version (patch/minor/major)
   - Generate changelog (using AI if OpenRouter API key is configured)
   - Build binaries for all platforms
   - Create and push a git tag
3. GitHub Actions will automatically:
   - Create a GitHub release
   - Upload binaries as release assets
   - Update README with installation instructions

For a dry run (without creating tags or pushing):
```bash
make release-dry-run
```

### Configuration

Create a `.env` file (see `.env.example`) to enable AI-powered changelog generation:
- `OPENROUTER_API_KEY`: Your OpenRouter API key for AI analysis
- `GITHUB_TOKEN`: Optional, for local release creation
- `GIT_REMOTE`: Git remote name (default: origin)
- `GIT_BRANCH`: Git branch name (default: master)

## Contributing

Contributions are welcome! Please ensure all tests pass and follow the existing code style.

## License

MIT License - see [LICENSE](LICENSE) file for details.

