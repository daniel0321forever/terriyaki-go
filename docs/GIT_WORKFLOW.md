# Git Workflow Guide: Working on Feature Branches

## ✅ What We Just Did

You were working on the `feature/elevenlabs` branch, but `main` had new updates. We successfully merged those updates into your feature branch.

### Steps Taken:
1. **Switched to feature branch**: `git checkout feature/elevenlabs`
2. **Merged main into feature**: `git merge origin/main`
3. **Resolved conflict**: Fixed merge conflict in `main.go` by keeping both changes
4. **Completed merge**: Committed the resolved conflict

### Result:
- ✅ Your voice changer feature is still intact
- ✅ All new changes from main are now in your branch
- ✅ Your branch is up-to-date with main

---

## 📚 Understanding the Situation

### The Problem:
```
main branch:     A---B---C---D---E (new updates)
                      \
feature branch:        F---G (your voice changer work)
```

When `main` gets updates while you're working on a feature branch, you need to bring those changes into your branch.

### The Solution:
```
main branch:     A---B---C---D---E
                      \           \
feature branch:        F---G-------M (merge commit)
```

---

## 🔄 Best Practices for Feature Branch Development

### 1. **Regularly Update Your Feature Branch**

**When to do it:**
- Before starting new work each day
- When you hear main has important updates
- Before creating a pull request

**How to do it:**
```bash
# Make sure you're on your feature branch
git checkout feature/elevenlabs

# Fetch latest changes from remote
git fetch origin

# Merge main into your feature branch
git merge origin/main
```

### 2. **Handle Merge Conflicts**

**What is a conflict?**
- When both branches modified the same file
- Git can't automatically decide which changes to keep
- You need to manually resolve it

**How to resolve:**
1. Git will mark conflicts in files with `<<<<<<<`, `=======`, `>>>>>>>`
2. Open the file and look for these markers
3. Keep the changes you need from both branches
4. Remove the conflict markers
5. Stage the file: `git add <filename>`
6. Complete the merge: `git commit`

**Example conflict:**
```go
// Your feature branch has:
router.POST("/api/v1/voice/convert", api.ConvertVoiceAPI)

// Main branch has:
router.PATCH("/api/v1/profile", api.UpdateProfileAPI)

// Resolved (keep both):
router.POST("/api/v1/voice/convert", api.ConvertVoiceAPI)
router.PATCH("/api/v1/profile", api.UpdateProfileAPI)
```

### 3. **Workflow Checklist**

**Starting a new feature:**
```bash
# 1. Make sure main is up-to-date
git checkout main
git pull origin main

# 2. Create feature branch from main
git checkout -b feature/your-feature-name

# 3. Start coding!
```

**During development:**
```bash
# Commit your work regularly
git add .
git commit -m "feat: add voice conversion feature"

# Push to remote
git push origin feature/your-feature-name
```

**When main has updates:**
```bash
# 1. Make sure your work is committed
git status  # Should be clean

# 2. Fetch latest from remote
git fetch origin

# 3. Merge main into your branch
git merge origin/main

# 4. Resolve any conflicts if needed
# 5. Test your code still works
# 6. Push updated branch
git push origin feature/your-feature-name
```

**When feature is complete:**
```bash
# 1. Make sure main is merged (as above)
# 2. Create pull request on GitHub/GitLab
# 3. After PR is approved and merged, delete local branch:
git checkout main
git pull origin main
git branch -d feature/your-feature-name
```

---

## 🎯 Common Scenarios

### Scenario 1: "I want the latest from main"
```bash
git checkout feature/elevenlabs
git fetch origin
git merge origin/main
# Resolve conflicts if any
git push origin feature/elevenlabs
```

### Scenario 2: "I made changes on main by mistake"
```bash
# Save your changes
git stash

# Switch to feature branch
git checkout feature/elevenlabs

# Apply your changes
git stash pop

# Commit to feature branch
git add .
git commit -m "feat: your changes"
```

### Scenario 3: "I want to start fresh from latest main"
```bash
# Save your feature branch work
git checkout feature/elevenlabs
git push origin feature/elevenlabs  # Backup first!

# Get latest main
git checkout main
git pull origin main

# Recreate feature branch from latest main
git checkout -b feature/elevenlabs-v2

# Cherry-pick your commits if needed
git cherry-pick <commit-hash>
```

---

## ⚠️ Important Rules

1. **Never force push to shared branches**
   - `git push --force` can destroy others' work
   - Only use on your own feature branches if absolutely necessary

2. **Always commit before merging**
   - Uncommitted changes can be lost
   - `git status` to check before merging

3. **Test after merging**
   - Make sure your code still works
   - Run tests if you have them

4. **Communicate with your team**
   - Let others know if you're working on shared files
   - Coordinate large merges

---

## 🆘 Emergency Commands

**Abort a merge:**
```bash
git merge --abort
```

**Undo last commit (keep changes):**
```bash
git reset --soft HEAD~1
```

**Undo last commit (discard changes):**
```bash
git reset --hard HEAD~1
```

**See what changed:**
```bash
git log --oneline --graph --all
git diff main..feature/elevenlabs
```

---

## 📖 Summary

**What happened today:**
- ✅ Merged latest main into feature/elevenlabs
- ✅ Resolved conflict in main.go
- ✅ Kept both: voice changer route + new routes from main
- ✅ Your feature is now up-to-date!

**Next steps:**
- Continue working on your feature branch
- Regularly merge main to stay updated
- When ready, create a pull request to merge back to main

**Remember:**
- Feature branches are for isolated work
- Regularly sync with main to avoid big conflicts
- Always test after merging
- Commit often, push regularly

---

## 🎓 Learning Resources

- **Git Basics**: https://git-scm.com/doc
- **Branching Strategy**: https://www.atlassian.com/git/tutorials/comparing-workflows
- **Resolving Conflicts**: https://git-scm.com/book/en/v2/Git-Tools-Advanced-Merging
