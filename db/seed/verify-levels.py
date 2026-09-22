#!/usr/bin/env python3
"""Đối chiếu bậc CEFR của từ vựng trong database với dữ liệu công bố.

Khẳng định "từ này thuộc bậc B1" là thứ nhìn bằng mắt không kiểm được, nên nó
phải kiểm được bằng lệnh. Script tải hai danh sách nguồn rồi báo mọi chỗ lệch.

    make verify-vocab

Nguồn:
  A1–B2  CEFR-J Wordlist Version 1.5 — Yukio Tono, Tokyo University of Foreign
         Studies. Dùng được cho nghiên cứu và thương mại, kèm trích dẫn.
  C1–C2  Octanove Vocabulary Profile C1/C2 v1.0 — Octanove Labs, CC BY-SA 4.0.
  Cả hai: https://github.com/openlanguageprofiles/olp-en-cefrj
"""

import csv, io, subprocess, sys

BASE = "https://raw.githubusercontent.com/openlanguageprofiles/olp-en-cefrj/master"
SOURCES = ("cefrj-vocabulary-profile-1.5.csv", "octanove-vocabulary-profile-c1c2-1.0.csv")
ORDER = ["A1", "A2", "B1", "B2", "C1", "C2"]


def load_reference() -> dict[str, str]:
    """Bậc thấp nhất thắng: từ xuất hiện ở nhiều nơi thì tính là học sớm nhất."""
    ref: dict[str, str] = {}
    for name in SOURCES:
        # curl thay vì urllib: Python cài từ python.org trên macOS thường
        # thiếu bộ CA và ngã ngay ở bước bắt tay TLS.
        body = subprocess.run(
            ["curl", "-fsSL", f"{BASE}/{name}"],
            capture_output=True, text=True, check=True,
        ).stdout
        for row in csv.DictReader(io.StringIO(body)):
            level = row["CEFR"].strip()
            if level not in ORDER:
                continue
            # Headword có thể gộp biến thể: "a.m./A.M./am/AM".
            for word in (w.strip().lower() for w in row["headword"].split("/")):
                if word and (word not in ref or ORDER.index(level) < ORDER.index(ref[word])):
                    ref[word] = level
    return ref


def load_database() -> list[tuple[str, str, str]]:
    query = """
        SELECT lower(v.word), c.level, l.slug
        FROM vocabulary_entries AS v
        JOIN lessons AS l ON l.id = v.lesson_id
        JOIN courses AS c ON c.id = l.course_id
        WHERE v.deleted_at IS NULL AND l.deleted_at IS NULL AND c.deleted_at IS NULL
    """
    out = subprocess.run(
        ["docker", "compose", "exec", "-T", "postgres",
         "psql", "-U", "lingora", "-d", "lingora", "-tAF\t", "-c", query],
        capture_output=True, text=True, check=True,
    ).stdout
    return [tuple(line.split("\t")) for line in out.splitlines() if line.strip()]


def main() -> int:
    reference = load_reference()
    entries = load_database()
    if not entries:
        print("Không có từ vựng nào trong database. Chạy `make seed` trước.")
        return 1

    problems = []
    for word, level, lesson in entries:
        actual = reference.get(word)
        if actual is None:
            problems.append((word, lesson, f"đặt {level}, không có trong dữ liệu nguồn"))
        elif actual != level:
            problems.append((word, lesson, f"đặt {level}, nguồn nói {actual}"))

    print(f"{len(entries)} từ, {len(entries) - len(problems)} khớp bậc với nguồn.")
    for word, lesson, why in problems:
        print(f"  ✗ {word:<20} bài {lesson:<24} {why}")
    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())
