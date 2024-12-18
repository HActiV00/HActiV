// 10-27 사이드바 가로 크기 변경

// @mui material 컴포넌트
import Drawer from "@mui/material/Drawer";
import { styled } from "@mui/material/styles";

// Drawer 컴포넌트를 스타일링한 커스텀 컴포넌트를 생성합니다.
export default styled(Drawer)(({ theme, ownerState }) => {
  // 테마와 ownerState에서 필요한 값을 가져옵니다.
  const { palette, boxShadows, transitions, breakpoints, functions } = theme;
  const { transparentSidenav, whiteSidenav, miniSidenav, darkMode } = ownerState;

  // 사이드바의 너비를 설정합니다.
  const sidebarWidth = 220; // 사이드바 가로 크기 설정 (기존 250에서 220으로 변경)
  const { transparent, gradients, white, background } = palette;
  const { xxl } = boxShadows;
  const { pxToRem, linearGradient } = functions;

  // 사이드바의 배경색을 설정합니다.
  let backgroundValue = darkMode
    ? background.sidenav // 다크 모드일 때의 배경색
    : linearGradient(gradients.dark.main, gradients.dark.state); // 일반 모드의 그라데이션 배경

  // 투명한 사이드바 설정일 경우 배경색을 투명으로 설정
  if (transparentSidenav) {
    backgroundValue = transparent.main;
  } else if (whiteSidenav) {
    // 흰색 사이드바 설정일 경우 배경색을 흰색으로 설정
    backgroundValue = white.main;
  }

  // miniSidenav가 false일 때의 스타일 설정 (사이드바가 확장된 상태)
  const drawerOpenStyles = () => ({
    background: backgroundValue,
    transform: "translateX(0)", // 사이드바를 원래 위치에 고정
    transition: transitions.create("transform", {
      easing: transitions.easing.sharp,
      duration: transitions.duration.shorter,
    }),

    [breakpoints.up("xl")]: {
      boxShadow: transparentSidenav ? "none" : xxl,
      marginBottom: transparentSidenav ? 0 : "inherit",
      left: "0",
      width: sidebarWidth, // 설정한 사이드바의 너비 적용
      transform: "translateX(0)", // 사이드바를 원래 위치에 고정
      transition: transitions.create(["width", "background-color"], {
        easing: transitions.easing.sharp,
        duration: transitions.duration.enteringScreen,
      }),
    },
  });

  // miniSidenav가 true일 때의 스타일 설정 (사이드바가 축소된 상태)
  const drawerCloseStyles = () => ({
    background: backgroundValue,
    transform: `translateX(${pxToRem(-320)})`, // 사이드바를 화면 밖으로 숨김
    transition: transitions.create("transform", {
      easing: transitions.easing.sharp,
      duration: transitions.duration.shorter,
    }),

    [breakpoints.up("xl")]: {
      boxShadow: transparentSidenav ? "none" : xxl,
      marginBottom: transparentSidenav ? 0 : "inherit",
      left: "0",
      width: pxToRem(96), // 축소된 사이드바의 너비 설정
      overflowX: "hidden", // 내용이 넘칠 경우 숨김 처리
      transform: "translateX(0)", // 사이드바를 원래 위치에 고정
      transition: transitions.create(["width", "background-color"], {
        easing: transitions.easing.sharp,
        duration: transitions.duration.shorter,
      }),
    },
  });

  return {
    "& .MuiDrawer-paper": {
      boxShadow: xxl, // 사이드바 그림자 설정
      border: "none", // 사이드바의 테두리 제거

      ...(miniSidenav ? drawerCloseStyles() : drawerOpenStyles()), // miniSidenav에 따라 다른 스타일 적용
    },
  };
});
