<h2> [HActiV-Docker(컨테이너 통합 버전)] </h2>

  1. 도구 설치
  ```bash
  glt clone https://github.com/HActiV00/HActiV.git
  cd HActiV-Tool
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

## [HActiV-Host(ubuntu 20.04/22.04/24.04)]

  1. 도구 설치

  ```bash
  glt clone https://github.com/HActiV00/HActiV.git
  
  cd HActiV
  ```

  2. 필수 빌드 도구 및 라이브러리 업데이트
  
  ```bash
  sudo apt-get update
  sudo apt install git make pkg-config build-essential libelf-dev libpcap-dev zlib1g-dev cmake
  ```
  
  3. eBPF 관련 도구 및 헤더 설치
  
  ```bash
  sudo apt-get install bpfcc-tools linux-headers-$(uname -r) python3
  sudo apt-get install bpfcc-tools linux-headers-$(uname -r) python3-bpfcc
  ```
  
  4. LLVM(+Clang) 설치
  
  ```bash
  sudo apt-get install -y wget gnupg
  wget https://apt.llvm.org/llvm.sh
  chmod +x llvm.sh
  sudo ./llvm.sh 13
  sudo apt-get update
  sudo apt-get install clang-13 libclang-13-dev llvm-13
  sudo update-alternatives --install /usr/bin/clang clang /usr/bin/clang-13 100
  sudo update-alternatives --install /usr/bin/llvm-config llvm-config /usr/bin/llvm-config-13 100
  ```
  
  5. git으로 BCC 설치
  
  ```bash
  # 1. Clone and enter BCC project
  git clone https://github.com/iovisor/bcc.git
  cd bcc
  
  # 2. Create build directory
  mkdir build
  cd build
  
  # 4. Run cmake to configure the build
  cmake -DLLVM_DIR=/usr/lib/llvm-13/cmake -DCMAKE_C_COMPILER=/usr/bin/clang-13 -DCMAKE_CXX_COMPILER=/usr/bin/clang++-13 ..
  
  # 5. Compile and install
  make
  sudo make install
  
  # 6. Cache update
  sudo ldconfig
  ```
  
  6. go 언어 설치

  ```bash
  sudo snap install go --classic
  ```
  
  7. HActiV Agent 빌드
  
  ```bash
  # [HActiV]Agent
  
  # etc - /etc/HActiV/ 디렉토리 아래에 덮어쓰기
  mv * /etc/HActiV/
  mv: cannot move 'rules' to '/etc/HActiV/rules': Directory not empty
  cd rules
  mv * /etc/HActiV/rules/
  
  # HActiV - Agent
  cd HActiV/cmd/
  make
  ```
  
  8. HActiV Setting.js 설정
  
  ```bash
  # HActiV 웹 url 사용 (/etc/HActiV/Setting.json)
  
  {
    "API": "your-secret-api-key",
    "HostMonitoring": "true",
    "LogLocation": "/etc/HActiV/logs",
    "Region": "Asia/Seoul",
    "RuleLocation": "/etc/HActiV/rules",
    "DataUrl": "https://hactiv-dev.run.goorm.site/api/dashboard/api/dashboard",
    "DataSend": "True",
    "RuleUrl": "https://hactiv-dev.run.goorm.site/api/dashboard/api/alert"
  }
  ```

  ```bash
  # 호스트 웹 url 사용  (/etc/HActiV/Setting.json)
  
  {
    "API": "your-secret-api-key",
    "HostMonitoring": "true",
    "LogLocation": "/etc/HActiV/logs",
    "Region": "Asia/Seoul",
    "RuleLocation": "/etc/HActiV/rules",
    "DataUrl": "http://hactiv-web-backend:8080/api/dashboard",
    "DataSend": "True",
    "RuleUrl": "http://hactiv-web-backend:8080/api/alert"
  }
  ```

  9. HActiV 실행
  ```bash
  ./HActiV {arg1} {arg2} {arg3} ... {arg4}
  ```
