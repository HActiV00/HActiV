FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    zip bison build-essential cmake flex git libedit-dev \
    python3 python3-setuptools liblzma-dev arping netperf iperf \
    libelf-dev llvm-11-dev libclang-11-dev linux-headers-generic \
    && apt-get clean

ENV DEBIAN_FRONTEND=dialog

RUN if ! command -v clang-11 &> /dev/null; then \
        apt-get install -y clang-11 llvm-11 libclang-11-dev && \
        update-alternatives --install /usr/bin/clang clang /usr/bin/clang-11 100 && \
        update-alternatives --install /usr/bin/clang++ clang++ /usr/bin/clang++-11 100; \
    fi

WORKDIR /tmp
RUN if [ ! -d "/usr/share/bcc" ]; then \
        git clone https://github.com/iovisor/bcc.git && \
        mkdir bcc/build && cd bcc/build && \
        cmake -DLLVM_DIR=/usr/lib/llvm-11/cmake .. && \
        make && make install; \
    fi

RUN mkdir -p /Go
WORKDIR /Go
RUN if [ ! -d "/usr/local/go" ]; then \
        apt-get install -y wget && \
        wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz && \
        tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz && \
        rm go1.23.4.linux-amd64.tar.gz; \
    fi

# HActiVAgent 폴더 복사 및 빌드
COPY HActiVAgent /HActiV
WORKDIR /HActiV/cmd
RUN export PATH=$PATH:/usr/local/go/bin && \
    if [ ! -f "/HActiV/cmd/HActiVAgent" ]; then \
        apt-get install -y libpcap-dev && make; \
    fi

RUN apt-get install -y linux-headers-$(uname -r)
#RUN apt-get install -y bpfcc-tools

# HActiV 폴더를 /etc/HActiV로 이동
COPY HActiV /etc/HActiV

# 작업 디렉토리 설정
WORKDIR /HActiV/cmd

CMD ["/bin/bash"]

