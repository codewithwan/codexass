from __future__ import annotations

from pathlib import Path
import sys


ROOT = Path(__file__).resolve().parent.parent
LIMIT = 250


def main() -> int:
    failures: list[str] = []

    for path in sorted(ROOT.rglob("*.go")):
        text = path.read_text(encoding="utf-8")
        line_count = len(text.splitlines())
        rel = path.relative_to(ROOT)

        if line_count > LIMIT:
            failures.append(f"{rel}: {line_count} lines > {LIMIT}")

        if " any" in text or "any " in text or "[]any" in text or "map[string]any" in text:
            failures.append(f"{rel}: contains banned any usage")

    if failures:
        print("Go guard check failed:")
        for failure in failures:
            print(f"- {failure}")
        return 1

    print("Go guard check passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
