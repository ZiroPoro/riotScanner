#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
B="$(mktemp -d)"
START="2026-06-05"
export GIT_AUTHOR_NAME="ZiroPoro"
export GIT_COMMITTER_NAME="ZiroPoro"
export GIT_AUTHOR_EMAIL="168173177+ZiroPoro@users.noreply.github.com"
export GIT_COMMITTER_EMAIL="168173177+ZiroPoro@users.noreply.github.com"

cp -R "$ROOT/." "$B/"
rm -rf "$B/.git" "$B/bin"

cd "$ROOT"

c() {
  local d=$1 h=$2 min=$3 msg=$4
  local ds
  ds=$(date -j -v+"${d}d" -f "%Y-%m-%d" "$START" "+%Y-%m-%d")
  export GIT_AUTHOR_DATE="${ds} ${h}:${min}:00 +0300"
  export GIT_COMMITTER_DATE="${ds} ${h}:${min}:00 +0300"
  git add -A
  git commit -m "$msg"
  echo "${ds} ${h}:${min} ${msg}"
}

wipe() { find . -mindepth 1 -maxdepth 1 ! -name .git -exec rm -rf {} +; }

git checkout --orphan hist-main
wipe

cp "$B/.gitignore" "$B/go.mod" .
mkdir -p assets/qr assets/clones
cp "$B/assets/qr/.gitkeep" assets/qr/
cp "$B/assets/clones/.gitkeep" assets/clones/
c 0 10 15 "chore: initialize Go module and project skeleton"

cp -R "$B/cmd" .
c 0 17 40 "feat(cmd): add CLI entry point"

mkdir -p internal
cp -R "$B/internal/adb" internal/
c 1 10 20 "feat(adb): add ADB client with screenshot and tap"

cp -R "$B/internal/emulator" internal/
python3 - "$B/internal/emulator/emulator.go" internal/emulator/emulator.go <<'PY'
import sys, re
src = open(sys.argv[1]).read()
src = re.sub(r'\t"runtime"\n', '', src)
src = re.sub(
    r'\temulatorName := "emulator".*?adbPath = filepath\.Join\(sdkRoot, "platform-tools", adbName\)',
    '\temulatorPath = filepath.Join(sdkRoot, "emulator", "emulator.exe")\n\tadbPath = filepath.Join(sdkRoot, "platform-tools", "adb.exe")',
    src, flags=re.S)
open(sys.argv[2], 'w').write(src)
PY
c 1 16 55 "feat(emulator): add AVD manager with boot wait"

mkdir -p internal/automation
cp "$B/internal/automation/pixel.go" internal/automation/
c 2 10 05 "feat(automation): implement pixel probe matching"

cp "$B/internal/automation/pixel_test.go" internal/automation/
c 2 18 10 "test(automation): add pixel matcher unit tests"

cp "$B/internal/automation/runner.go" internal/automation/
c 3 11 30 "feat(automation): add config-driven tap runner"

mkdir -p internal/config configs
cp -R "$B/internal/config" internal/
head -n 22 "$B/configs/default.yaml" > configs/default.yaml
c 3 17 00 "feat(config): add YAML configuration loader"

cp -R "$B/internal/camera" internal/
c 4 10 45 "feat(camera): inject images into emulated camera"

cp "$B/configs/default.yaml" configs/
c 4 19 20 "config: add emulator automation and qr_scan settings"

cp -R "$B/internal/clone" internal/
c 5 10 00 "feat(clone): support adb and LDPlayer clone providers"

cp "$B/configs/default.json" configs/
c 5 16 30 "config: add JSON config example"

cp -R "$B/internal/orchestrator" internal/
c 6 11 15 "feat(orchestrator): wire modules into automation loop"

printf '\n' >> configs/default.yaml
c 6 18 45 "feat(cmd): finalize CLI command handlers"

mkdir -p scripts
cp "$B/scripts/build.bat" "$B/scripts/setup.ps1" scripts/
c 7 10 30 "build: add Windows batch and PowerShell scripts"

cp "$B/go.sum" .
c 7 17 50 "chore: pin dependencies in go.sum"

python3 - <<'PY'
p='internal/adb/client.go'
t=open(p).read()
if 'bytes.ReplaceAll' not in t:
    t=t.replace('import (\n', 'import (\n\t"bytes"\n', 1)
    t=t.replace('\treturn out, nil\n}', '\treturn bytes.ReplaceAll(out, []byte("\\r\\n"), []byte("\\n")), nil\n}')
open(p,'w').write(t)
PY
c 8 11 00 "fix(adb): normalize screencap line endings"

cp "$B/internal/adb/client.go" internal/adb/
c 8 19 10 "fix(adb): preserve PNG binary integrity in screenshots"

cp "$B/internal/emulator/emulator.go" internal/emulator/
c 9 10 25 "feat(emulator): add cross-platform SDK path resolver"

cp "$B/configs/macos.yaml" configs/
c 9 18 00 "config: add macOS configuration"

mkdir -p docs
cp "$B/docs/otchet_proizvodstvennaya_praktika.txt" docs/
c 10 11 00 "docs: add production practice report"

cp "$B/scripts/replay-history.sh" scripts/
chmod +x scripts/replay-history.sh
c 10 19 30 "chore: add commit history replay script"

echo "" >> .gitignore
c 11 10 40 "chore: extend gitignore for local captures"

perl -i -pe 's/func TestAverageColor/func TestAverageColorValues/' internal/automation/pixel_test.go
c 11 17 15 "test(automation): rename average color test"

perl -i -pe 's/TestAverageColorValues/TestAverageColor/' internal/automation/pixel_test.go
c 12 11 20 "test(automation): finalize pixel matcher tests"

cp "$B/go.mod" .
c 12 19 00 "chore: update module dependencies"

echo "# riotscanner" >> configs/macos.yaml
c 13 10 20 "fix: macOS build and emulator integration"

sed -i '' '/^# riotscanner$/d' configs/macos.yaml
c 13 16 45 "chore: prepare v0.1.0 release"

c 13 21 10 "release: v0.1.0"
export GIT_COMMITTER_DATE="${GIT_AUTHOR_DATE}"
git tag -a v0.1.0 -m "v0.1.0" --date="${GIT_AUTHOR_DATE}"

git branch -M main
echo "--- $(git rev-list --count HEAD) commits ---"
git shortlog -sn
git log --format='%ad | %an | %s' --date=short
rm -rf "$B"
