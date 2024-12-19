<h2> HActiV-Tool(컨테이너 통합 버전) </h2>



https://github.com/user-attachments/assets/e825523a-5540-4646-a6d9-4df00f5df6cc


  1. 도구 설치
  ```bash
  git clone https://github.com/HActiV00/HActiV.git
  cd HActiV/1.HActiV-Tool/
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
  ./HActiV {arg1} {arg2} {arg3} ... {argN}
  ```
