## [HActiV-Web(Localhost)]

  1. HActiV-Web 설치
  ```bash
  glt clone https://github.com/HActiV00/HActiV.git
  mv -f HActiV-Web/* HActiV-Tool/
  cd HActiV-Tool  
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
