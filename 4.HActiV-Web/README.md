[  ## [4.HActiV-Web(호스트 웹 버전)]](https://github.com/HActiV00/HActiV/tree/main/HActiV-Web)

  1. HActiV-Web 설치
  ```bash
  glt clone https://github.com/HActiV00/HActiV.git
  mv -f 4.HActiV-Web/* 1.HActiV-Tool/
  cd 1.HActiV-Tool  
  ```

  2. HActiV-web 빌드
  ```bash
  docker-compose build
  docker-compose up
  ```

  3. HActiV-web 접속
  ```bash
  - frontend (http://localhost:3000)
  - backend (http://localhost:8080)
  ```
