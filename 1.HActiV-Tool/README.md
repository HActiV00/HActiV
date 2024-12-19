<h2> [HActiV-Tool(컨테이너 통합 버전)] </h2>

  1. 도구 설치
  ```bash
  glt clone https://github.com/HActiV00/HActiV.git
  cd 1.HActiV-Tool
  ```
  2. 웹 & 도구 빌드
  ```bash
  docker-compose build
  docker-compose up
  ```

  3. 도구 접속
  ```bash
  docker exec -it  HActiV /bin/bash
  ```

  4. 도구 실행
  ```bash
  ./HActiV {arg1} {arg2} {arg3} ... {arg4}
  ```
