import React, { useEffect } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import SystemCallDetail from "./details/SystemCallDetail";
import FileOpenDetail from "./details/FileOpenDetail";
import NetworkTrafficDetail from "./details/NetworkTrafficDetail";
import MemoryDetail from "./details/MemoryDetail";
import LogFileEventDetail from "./details/LogFileEventDetail"; // 통합된 로그 파일 이벤트 페이지

export default function Detail() {
  const [searchParams, setSearchParams] = useSearchParams();
  const tool = searchParams.get("tool");
  const navigate = useNavigate();

  useEffect(() => {
    if (tool === "log_file_open" || tool === "log_file_delete") {
      // 파라미터를 log_file_event로 변경
      setSearchParams({ tool: "log_file_event" });
    }
	if (tool === "file_open" || tool === "delete") {
      setSearchParams({ tool: "file_event" });
    }
  }, [tool, setSearchParams]);

  const renderToolDetail = () => {
    switch (tool) {
      case "Systemcall":
        return <SystemCallDetail />;
      case "file_event":
        return <FileOpenDetail />;
      case "Network_traffic":
        return <NetworkTrafficDetail />;
      case "Memory":
        return <MemoryDetail />;
      case "log_file_event": // 통합된 로그 파일 이벤트
        return <LogFileEventDetail />;
      default:
        return <p>유효한 도구가 선택되지 않았습니다.</p>;
    }
  };

  return (
    <div>
      {renderToolDetail()}
    </div>
  );
}
