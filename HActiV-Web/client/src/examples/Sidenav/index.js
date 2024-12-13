import { useEffect } from "react";

// react-router-dom 컴포넌트
import { useLocation, NavLink } from "react-router-dom";

// prop-types는 props의 타입 체크 라이브러리입니다.
import PropTypes from "prop-types";

// @mui material 컴포넌트
import List from "@mui/material/List";
import Divider from "@mui/material/Divider";
import Link from "@mui/material/Link";
import Icon from "@mui/material/Icon";

// Material Dashboard 2 React 컴포넌트
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";

// Material Dashboard 2 React 예제 컴포넌트
import SidenavCollapse from "examples/Sidenav/SidenavCollapse";

// Sidenav의 커스텀 스타일
import SidenavRoot from "examples/Sidenav/SidenavRoot";
import sidenavLogoLabel from "examples/Sidenav/styles/sidenav";

// Material Dashboard 2 React 컨텍스트
import {
  useMaterialUIController,
  setMiniSidenav,
  setTransparentSidenav,
  setWhiteSidenav,
} from "context";

function Sidenav({ color, brand, brandName, routes, ...rest }) {
  // 컨텍스트에서 컨트롤러와 dispatch 함수 가져오기
  const [controller, dispatch] = useMaterialUIController();
  const { miniSidenav, transparentSidenav, whiteSidenav, darkMode } = controller;

  // 현재 위치 정보를 가져와서 경로를 설정
  const location = useLocation();
  const collapseName = location.pathname.replace("/", "");

  // 텍스트 색상 설정
  let textColor = "white";
  if (transparentSidenav || (whiteSidenav && !darkMode)) {
    textColor = "dark";
  } else if (whiteSidenav && darkMode) {
    textColor = "inherit";
  }

  // 사이드바를 닫는 함수
  const closeSidenav = () => setMiniSidenav(dispatch, true);

  useEffect(() => {
    // 사이드 내비게이션의 mini 상태를 설정하는 함수
    function handleMiniSidenav() {
      setMiniSidenav(dispatch, window.innerWidth < 1200);
      setTransparentSidenav(dispatch, window.innerWidth < 1200 ? false : transparentSidenav);
      setWhiteSidenav(dispatch, window.innerWidth < 1200 ? false : whiteSidenav);
    }

    // 창 크기 조정 시 handleMiniSidenav 함수를 호출하는 이벤트 리스너
    window.addEventListener("resize", handleMiniSidenav);

    // 초기 값으로 상태를 설정하기 위해 handleMiniSidenav 함수 호출
    handleMiniSidenav();

    // 컴포넌트 언마운트 시 이벤트 리스너 제거
    return () => window.removeEventListener("resize", handleMiniSidenav);
  }, [dispatch, location, transparentSidenav, whiteSidenav]);

  // routes.js의 모든 경로를 렌더링합니다 (Sidenav에 표시되는 모든 항목)
  const renderRoutes = routes.map(({ type, name, icon, title, noCollapse, key, href, route }) => {
    let returnValue;

    // 'collapse' 타입의 항목 처리
    if (type === "collapse") {
      returnValue = href ? (
        <Link
          href={href}
          key={key}
          target="_blank"
          rel="noreferrer"
          sx={{ textDecoration: "none" }}
        >
          <SidenavCollapse
            name={name}
            icon={icon}
            active={key === collapseName}
            noCollapse={noCollapse}
          />
        </Link>
      ) : (
        <NavLink key={key} to={route}>
          <SidenavCollapse name={name} icon={icon} active={key === collapseName} />
        </NavLink>
      );
    } 
    // 'title' 타입의 항목 처리
    else if (type === "title") {
      returnValue = (
        <MDTypography
          key={key}
          color={textColor}
          display="block"
          variant="caption"
          fontWeight="bold"
          textTransform="uppercase"
          pl={3}
          mt={2}
          mb={1}
          ml={1}
        >
          {title}
        </MDTypography>
      );
    } 
    // 'divider' 타입의 항목 처리
    else if (type === "divider") {
      returnValue = (
        <Divider
          key={key}
          light={
            (!darkMode && !whiteSidenav && !transparentSidenav) ||
            (darkMode && !transparentSidenav && whiteSidenav)
          }
        />
      );
    }

    return returnValue;
  });

  return (
    <SidenavRoot
      {...rest}
      variant="permanent"
      ownerState={{ transparentSidenav, whiteSidenav, miniSidenav, darkMode }}
    >
      {/* 상단 브랜드 로고 및 닫기 아이콘 */}
      <MDBox pt={3} pb={1} px={5} textAlign="center">
        <MDBox
          display={{ xs: "block", xl: "none" }} // 작은 화면에서 사이드바 가운데 맞춤
          position="absolute"
          top={0}
          right={0}
          p={1.625} 
          onClick={closeSidenav}
          sx={{ cursor: "pointer" }}
        >
          <MDTypography variant="h6" color="secondary">
            <Icon sx={{ fontWeight: "bold" }}>close</Icon>
          </MDTypography>
        </MDBox>
        {/* 브랜드 로고와 이름 */}
        <MDBox component={NavLink} to="/" display="flex" alignItems="center">
          {brand && <MDBox component="img" src={brand} alt="Brand" width="2rem" />}
          <MDBox
            width={!brandName && "100%"}
            sx={(theme) => sidenavLogoLabel(theme, { miniSidenav })}
          >
            <MDTypography component="h6" variant="button" fontWeight="medium" color={textColor}>
              {brandName}
            </MDTypography>
          </MDBox>
        </MDBox>
      </MDBox>

      {/* 구분선 */}
      <Divider
        light={
          (!darkMode && !whiteSidenav && !transparentSidenav) ||
          (darkMode && !transparentSidenav && whiteSidenav)
        }
      />

      {/* 사이드바 내 경로 리스트 */}
      <List>{renderRoutes}</List>

    </SidenavRoot>
  );
}

// Sidenav의 기본 props 설정
Sidenav.defaultProps = {
  color: "info",
  brand: "",
};

// Sidenav의 props 타입 검사
Sidenav.propTypes = {
  color: PropTypes.oneOf(["primary", "secondary", "info", "success", "warning", "error", "dark"]), // 색상 옵션
  brand: PropTypes.string, // 브랜드 로고 이미지 URL
  brandName: PropTypes.string.isRequired, // 브랜드 이름
  routes: PropTypes.arrayOf(PropTypes.object).isRequired, // 사이드바 경로 배열
};

export default Sidenav;
