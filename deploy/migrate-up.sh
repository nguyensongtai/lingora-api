#!/bin/sh
# Áp dụng mọi migration còn thiếu. Railway chạy lệnh này trước mỗi lần deploy,
# trong một container riêng; lỗi thì deploy dừng lại và bản đang chạy giữ
# nguyên. golang-migrate khoá bằng advisory lock của Postgres, nên hai lần
# deploy chồng nhau không chạy cùng một migration hai lần.
set -eu
: "${DATABASE_URL:?DATABASE_URL chưa được đặt}"
exec migrate -path /migrations -database "$DATABASE_URL" up
