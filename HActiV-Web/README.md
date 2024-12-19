##  [HActiV-Web(ubuntu 20.04/22.04/24.04)]

  1. Docker 설치

  ```bash
  sudo apt-get update
  sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
  sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
  sudo apt-get update
  sudo apt-get install -y docker-ce
  ```

  2. Docker compose 설치
  ```bash
  apt install docker-compose
  ```

  3. Docker 서비스 실행 및 상태 확인
  ```bash
  sudo systemctl start docker
  sudo systemctl enable docker
  sudo systemctl status docker
  ```

  4. HActiV-web 빌드
  ```bash
  docker-compose build
  docker-compose up
  ```

  5. HActiV-web 접속
  ```bash
  - frontend (http://localhost:3000)
  - backend (http://localhost:8080)
  ```