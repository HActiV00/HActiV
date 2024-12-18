// profile/components/ScheduleCard.js

import React, { useState } from "react";
import Card from "@mui/material/Card";
import MDBox from "components/MDBox";
import MDTypography from "components/MDTypography";
import TextField from "@mui/material/TextField";
import Calendar from "react-calendar";
import "react-calendar/dist/Calendar.css";

function ScheduleCard() {
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [notes, setNotes] = useState({});

  // 날짜 선택 시 호출되는 함수
  const handleDateChange = (date) => {
    setSelectedDate(date);
  };

  // 메모 입력 시 호출되는 함수
  const handleNoteChange = (event) => {
    const newNotes = { ...notes, [selectedDate.toDateString()]: event.target.value };
    setNotes(newNotes);
  };

  return (
    <Card>
      <MDBox p={2}>
        <MDTypography variant="h6" fontWeight="medium">
          일정 관리
        </MDTypography>

        <MDBox mt={2} display="flex" flexDirection="column" alignItems="center">
          <Calendar
            onChange={handleDateChange}
            value={selectedDate}
            formatDay={(locale, date) => date.getDate().toString()}
            style={{ width: "300px", height: "300px" }} // Calendar 크기 고정
          />

          <MDBox mt={2} width="100%">
            <MDTypography variant="body2" fontWeight="medium">
              선택된 날짜: {selectedDate.toLocaleDateString()}
            </MDTypography>

            <TextField
              aria-label="메모 입력"
              multiline
              minRows={3}
              placeholder="메모를 입력하세요"
              value={notes[selectedDate.toDateString()] || ""}
              onChange={handleNoteChange}
              fullWidth
              style={{ marginTop: "8px" }}
            />
          </MDBox>
        </MDBox>
      </MDBox>
    </Card>
  );
}

export default ScheduleCard;
