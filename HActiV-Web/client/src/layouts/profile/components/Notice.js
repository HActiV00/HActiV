import { useNavigate } from "react-router-dom";
import Card from "@mui/material/Card";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import MDButton from "components/MDButton";

const sampleAlerts = [
  { label: "Success Alert", message: "A success alert example." },
  { label: "Error Alert", message: "An error alert example." },
  { label: "Warning Alert", message: "A warning alert example." },
  { label: "Another Success Alert", message: "Another example of success alert." },
  { label: "Critical Error", message: "This is a critical error alert." },
];

function Notice() {
  const navigate = useNavigate();

  const handleViewMore = () => {
    navigate("/notifications");
  };

  return (
    <Card sx={{ p: 3, position: "relative" }}>
      {/* 상세보기 버튼 */}
      <MDBox position="absolute" top={16} right={16}>
        <MDButton variant="text" color="dark" size="small" onClick={handleViewMore}>
          상세보기
        </MDButton>
      </MDBox>

      {/* 공지 사항 본문 */}
      <MDBox mb={2}>
        <MDTypography variant="h6" fontWeight="medium">
          공지 사항
        </MDTypography>
        <MDTypography variant="body2" color="textSecondary">
          최신 공지 사항을 확인하세요.
        </MDTypography>
      </MDBox>

      {/* 공지 목록 */}
      {sampleAlerts.slice(0, 4).map((alert, index) => (
        <MDBox key={index} mb={1}>
          <MDTypography variant="body2" fontWeight="medium">
            {alert.label}
          </MDTypography>
          <MDTypography variant="caption" color="text">
            {alert.message}
          </MDTypography>
        </MDBox>
      ))}
    </Card>
  );
}

export default Notice;
