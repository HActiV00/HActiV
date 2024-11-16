# HActiV-agent

HActiV는 클라우드 및 컨테이너 환경에서의 사용자 행위를 실시간으로 모니터링하고 로그를 수집하는 오픈 소스 보안, 관리 도구입니다.

## 설치 방법

에이전트를 설치하려면 다음 명령어를 실행하세요:

```bash
git clone {}
cd HActiV-agent
sudo go build .
```

## 사용법

에이전트를 실행하려면 다음 명령어를 사용하세요:

```bash
sudo go run main.go <function_number(s)>
```

## 옵션 설명

1. 시스템 메트릭스 모니터링

2. execv 시스템 콜 모니터링

3. 파일 삭제 모니터링

4. 메모리 모니터링

5. 네트워크 모니터링

6. 파일 열람 모니터링

### 예시:
```bash
sudo go run main.go 1
sudo go run main.go 1 6
```

### 기본 설정파일:
```bash
/etc/HActiV
```
