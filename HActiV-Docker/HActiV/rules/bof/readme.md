# Buffer Overflow Attack 정책

메모리 이벤트와 파일 실행, 열람 이벤트를 기반으로 버퍼 오버플로우 및 관련 보안 위협을 탐지하기 위한 정책을 포함하고 있습니다.

## 실험 환경
Ubuntu 20.04

## 사전 준비
```bash
sudo apt update
sudo apt install gcc python3
```

## bof_test.c
```c
#include <stdio.h>
#include <string.h>

void vulnerable_function(char *input) {
    char buffer[64];
    strcpy(buffer, input); // 취약점: 경계 검사 없음
    printf("Buffer: %s\n", buffer);
}

int main(int argc, char *argv[]) {
    if (argc < 2) {
        printf("Usage: %s <input>\n", argv[0]);
        return 1;
    }
    vulnerable_function(argv[1]);
    return 0;
}

```

## HActiV 실행
```bash
./HActiV 1 2 3 4 5 6
```

## 보안 실험 설정 및 테스트
```bash
echo 0 > /proc/sys/kernel/randomize_va_space
gcc -o bof_test -fno-stack-protector -m64 -z execstack bof_test.c

./bof_test $(python3 -c "print('A' * 300)")
```


