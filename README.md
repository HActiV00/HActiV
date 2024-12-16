
1. docker container 사용의 경우

  //도구 설치
  cd HActiV
  sudo docker build -t hactiv-image .

  //웹 & 도구 빌드
  docker-compose build
  docker-compose up

  // 도구 접속
  sudo docker run --name HActiV -it --privileged \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -v /proc:/proc:ro \
      -v /sys/kernel:/sys/kernel:ro \
      hactiv-image
