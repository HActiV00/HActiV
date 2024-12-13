// profile/components/Header/index.js
import Card from "@mui/material/Card";
import Grid from "@mui/material/Grid";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import Agent from "../Agent"; // 상위 경로로 수정
import Notice from "../Notice"; // 상위 경로로 수정
import Policy from "../Policy"; // 상위 경로로 수정

function Header() {
  return (
    <MDBox position="relative" mb={5}>
      <Card sx={{ mt: 3, mx: 3, py: 2, px: 2 }}>
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Agent /> {/* 에이전트 관리 카드 */}
          </Grid>
          <Grid item xs={12} md={6}>
            <Notice /> {/* 공지사항 카드 */}
          </Grid>
          <Grid item xs={12} md={6}>
            <Policy /> {/* 정책 카드 */}
          </Grid>
          {/* 추가 카드들을 더 배치하려면 여기서 추가 */}
        </Grid>
      </Card>
    </MDBox>
  );
}

export default Header;
