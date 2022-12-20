#!/bin/sh 
# tệp này được chạy bởi /bin/sh vì chúng ta đang sử dụng alpine image => bash shell ko khả dụng

# đảm bảo rằng tập lệnh sẽ thoát luôn nếu trả về  trạng thái khác zero status
set -e

echo "start the app"
exec "$@"
