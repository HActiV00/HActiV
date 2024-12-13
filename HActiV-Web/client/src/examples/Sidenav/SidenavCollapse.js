// prop-types는 props의 타입을 검사하는 라이브러리입니다.
import PropTypes from "prop-types";

// @mui material 컴포넌트
import ListItem from "@mui/material/ListItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import Icon from "@mui/material/Icon";

// Material Dashboard 2 React 컴포넌트
import MDBox from "components/MDBox";

// SidenavCollapse의 커스텀 스타일
import {
  collapseItem,
  collapseIconBox,
  collapseIcon,
  collapseText,
} from "examples/Sidenav/styles/sidenavCollapse";

// Material Dashboard 2 React 컨텍스트
import { useMaterialUIController } from "context";

function SidenavCollapse({ icon, name, active, ...rest }) {
  // 컨텍스트에서 상태 값 가져오기
  const [controller] = useMaterialUIController();
  const { miniSidenav, transparentSidenav, whiteSidenav, darkMode, sidenavColor } = controller;

  return (
    <ListItem component="li">
      <MDBox
        {...rest}
        sx={(theme) =>
          collapseItem(theme, {
            active,
            transparentSidenav,
            whiteSidenav,
            darkMode,
            sidenavColor,
          })
        }
      >
        {/* 아이콘 영역 */}
        <ListItemIcon
          sx={(theme) =>
            collapseIconBox(theme, { transparentSidenav, whiteSidenav, darkMode, active })
          }
        >
          {typeof icon === "string" ? (
            // icon이 문자열일 경우 Material Icon을 렌더링
            <Icon sx={(theme) => collapseIcon(theme, { active })}>{icon}</Icon>
          ) : (
            // 그렇지 않을 경우 커스텀 아이콘을 직접 렌더링
            icon
          )}
        </ListItemIcon>

        {/* 아이템 이름 영역 */}
        <ListItemText
          primary={name}
          sx={(theme) =>
            collapseText(theme, {
              miniSidenav,
              transparentSidenav,
              whiteSidenav,
              active,
            })
          }
        />
      </MDBox>
    </ListItem>
  );
}

// SidenavCollapse의 props에 대한 기본값 설정
SidenavCollapse.defaultProps = {
  active: false,
};

// SidenavCollapse의 props 타입 검사
SidenavCollapse.propTypes = {
  icon: PropTypes.node.isRequired, // icon은 노드 타입이어야 합니다.
  name: PropTypes.string.isRequired, // name은 문자열 타입이어야 합니다.
  active: PropTypes.bool, // active는 불리언 타입이어야 합니다.
};

export default SidenavCollapse;
