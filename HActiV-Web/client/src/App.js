import { useEffect } from "react";
import { Routes, Route, Navigate, useLocation } from "react-router-dom";
import { ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import Icon from "@mui/material/Icon";

import MDBox from "components/MDBox";
import Sidebar from "./Sidebar"; // 새로 만든 Sidebar 컴포넌트
import Configurator from "examples/Configurator";

import theme from "assets/theme";
import themeDark from "assets/theme-dark";

import { useMaterialUIController, setOpenConfigurator } from "context";
import AgentControl from "layouts/AgentControl"; 
import HomePage from "layouts/marketing"; 
import Dashboard from "layouts/dashboard"; 
import Notification from "layouts/notifications";
import EventAlert from "layouts/eventalert";
import Mypage from "layouts/profile";
import Login from "layouts/authentication/sign-in";
import Register from "layouts/authentication/sign-up";
import AccountSettings from "layouts/AccountSettings"; // 계정 설정 페이지
import GroupInfo from "layouts/GroupInfo";
import Detail from "layouts/dashboard/detail"; // Detail 컴포넌트로 변경
import Investigation from "layouts/investigation";

export default function App() {
  const [controller, dispatch] = useMaterialUIController();
  const { layout, openConfigurator, darkMode } = controller;
  const { pathname } = useLocation();

  useEffect(() => {
    document.documentElement.scrollTop = 0;
    document.scrollingElement.scrollTop = 0;

    // 페이지 제목 설정
    if (pathname === "/") {
      document.title = "HActiV - HomePage"; // 홈페이지 제목
    } else if (pathname === "/dashboard") {
      document.title = "HActiV - Dashboard"; // 대시보드 제목
    }
  }, [pathname]);

  const handleConfiguratorOpen = () => setOpenConfigurator(dispatch, !openConfigurator);

  const configsButton = (
    <MDBox
      display="flex"
      justifyContent="center"
      alignItems="center"
      width="3.25rem"
      height="3.25rem"
      bgColor="white"
      shadow="sm"
      borderRadius="50%"
      position="fixed"
      right="2rem"
      bottom="2rem"
      zIndex={99}
      color="dark"
      sx={{ cursor: "pointer" }}
      onClick={handleConfiguratorOpen}
    >
      <Icon fontSize="small" color="inherit">
        settings
      </Icon>
    </MDBox>
  );

  // 사이드바를 홈페이지, 로그인, 회원가입 페이지에서는 렌더링하지 않음
  const shouldRenderSidebar = layout === "dashboard" && pathname !== "/" && pathname !== "/dashboard/sign-in" && pathname !== "/dashboard/sign-up";

  return (
    <ThemeProvider theme={darkMode ? themeDark : theme}>
      <CssBaseline />
      {shouldRenderSidebar && <Sidebar />} {/* 조건에 따라 사이드바 렌더링 */}
      <Configurator />
      {configsButton}
      <Routes>
        <Route path="/" element={<HomePage />} /> {/* 홈페이지 */}
        <Route path="/dashboard" element={<Dashboard />} /> {/* 대시보드 */}
        <Route path="/dashboard/notifications" element={<Notification />} /> {/* 공지사항 */}
		<Route path="/dashboard/investigation" element={<Investigation />} /> {/* 공지사항 */}
        <Route path="/dashboard/eventalert" element={<EventAlert />} /> {/* 이벤트 알림 */}
        <Route path="/dashboard/mypage" element={<Mypage />} /> {/* 마이페이지 */}
        <Route path="/dashboard/sign-in" element={<Login />} /> {/* 로그인 */}
        <Route path="/dashboard/sign-up" element={<Register />} /> {/* 회원가입 */}
        <Route path="/dashboard/agent-control" element={<AgentControl />} /> {/* 에이전트 제어 */}
        <Route path="/dashboard/mypage/account-setting" element={<AccountSettings />} /> {/* 계정 설정 페이지 */}
        <Route path="/group-info" element={<GroupInfo />} />
        <Route path="/dashboard/detail" element={<Detail />} /> {/* Detail 페이지 라우트 */}
        <Route path="*" element={<Navigate to="/" />} /> {/* 잘못된 경로는 홈페이지로 리다이렉트 */}
      </Routes>
    </ThemeProvider>
  );
}
