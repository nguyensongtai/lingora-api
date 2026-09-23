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


def load_reference() -> tuple[dict[str, str], dict[str, dict[str, str]]]:
    """Bậc thấp nhất thắng: từ xuất hiện ở nhiều nơi thì tính là học sớm nhất.

    Trả kèm bậc theo từng từ loại. Database không lưu từ loại, nên một từ khớp
    bậc có thể chỉ khớp ở nghĩa KHÁC nghĩa đang dạy: bite từng qua kiểm nhờ
    danh từ A2 trong khi bài dạy động từ, vốn là B1.
    """
    ref: dict[str, str] = {}
    senses: dict[str, dict[str, str]] = {}
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
            pos = row["pos"].strip()
            for word in (w.strip().lower() for w in row["headword"].split("/")):
                if not word:
                    continue
                if word not in ref or ORDER.index(level) < ORDER.index(ref[word]):
                    ref[word] = level
                by_pos = senses.setdefault(word, {})
                if pos not in by_pos or ORDER.index(level) < ORDER.index(by_pos[pos]):
                    by_pos[pos] = level
    return ref, senses


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
    reference, senses = load_reference()
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

    # Không tính là lỗi — máy không biết bài dạy nghĩa nào. Nhưng người soát
    # phải nhìn nghĩa tiếng Việt và câu ví dụ để chắc bài dạy đúng từ loại này.
    ambiguous = []
    for word, level, lesson in entries:
        by_pos = senses.get(word, {})
        if len(set(by_pos.values())) > 1:
            matching = sorted(pos for pos, lv in by_pos.items() if lv == level)
            ambiguous.append((word, lesson, level, matching, by_pos))
    if ambiguous:
        print(f"\n{len(ambiguous)} từ có bậc khác nhau theo từ loại — bài phải dạy đúng nghĩa ghi bên phải:")
        for word, lesson, level, matching, by_pos in ambiguous:
            others = ", ".join(f"{pos} {lv}" for pos, lv in sorted(by_pos.items()))
            print(f"  ? {word:<20} bài {lesson:<24} {level} chỉ đúng khi dạy {'/'.join(matching) or '—'}  ({others})")
    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())
