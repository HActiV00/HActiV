
1. docker container 사용의 경우

  //도구 설치
  
  glt clonev https://github.com/HActiV00/HActiV.git
  
  cd HActiV

  //웹 & 도구 빌드
  
  docker-compose build
  
  docker-compose up


  // 도구 접속
  
  docker exec -it  HActiV /bin/bash
