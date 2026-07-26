# Workflow Conventions

## Branch
- Default branch: `master`

## Git Worktrees
All feature work uses git worktrees for isolation. See `docs/superpowers/plans/` per feature.

Process:
1. `master` branch is ground truth — always up-to-date with origin
2. Each feature gets a worktree: `../tv-<feature>/`
3. Worktrees are temporary — deleted after merge to master
4. Never commit directly to master (except trivial fixes)

## GitHub
- Account: mutedfalcontv
- Repo: mutedfalcontv/tv
