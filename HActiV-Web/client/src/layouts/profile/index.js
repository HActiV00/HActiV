// profile/index.js

import React, { useEffect, useState } from "react";
import Grid from "@mui/material/Grid";
import MDBox from "components/MDBox";
import DashboardLayout from "examples/LayoutContainers/DashboardLayout";
import DashboardNavbar from "examples/Navbars/DashboardNavbar";
import Footer from "examples/Footer";
import UserProfileCard from "./components/UserProfileCard";
import TeamSpaceCard from "./components/TeamSpaceCard";
import AgentCard from "./components/Agent";
import NoticeCard from "./components/Notice";
import PolicyCard from "./components/Policy";
import ScheduleCard from "./components/ScheduleCard"; 
import EventsAndAlertsCard from "./components/EventsAndAlertsCard"; // 이벤트 카드 임포트

function Profile() {
  const [profile, setProfile] = useState({});
  const [containers, setContainers] = useState([]);
  const [events, setEvents] = useState([
    { title: "Test Event", severity: "HIGH", description: "This is a test event" },
  ]); // 테스트용 기본 이벤트 추가

  useEffect(() => {
    async function fetchData() {
      const profileData = await fetch("/api/user").then((res) => res.json());
      const containerData = await fetch("/api/containers").then((res) => res.json());
      const eventsData = await fetch("/api/events").then((res) => res.json());

      setProfile(profileData);
      setContainers(containerData);
      setEvents(eventsData);
	  //setNotifications(notificationsData); 
    }

    fetchData();
  }, []);

  return (
    <DashboardLayout>
      <DashboardNavbar />
      
      {/* 메인 컨텐츠 박스 */}
      <MDBox mt={4} mb={3}>
        <Grid container spacing={3}>
          {/* 좌측 중앙 카드들 */}
          <Grid item xs={12} md={8} lg={9}>
            <Grid container spacing={3}>
              <Grid item xs={12}>
                <AgentCard />
              </Grid>
              {/* EventsAndAlertsCard 부분 */}
              <Grid item xs={12} md={6}>
                <EventsAndAlertsCard events={events} />
              </Grid>
			  <Grid item xs={12} md={6}>
                   <ScheduleCard /> {/* 기존 ContainerMetricsCard를 ScheduleCard로 변경 */}
              </Grid>

              <Grid item xs={12} md={6}>
                <NoticeCard />
              </Grid>
              <Grid item xs={12} md={6}>
                <PolicyCard />
              </Grid>
            </Grid>
          </Grid>

          {/* 우측 사이드바 카드들 */}
          <Grid item xs={12} md={4} lg={3}>
            <MDBox mb={3}>
              <UserProfileCard profile={profile} />
            </MDBox>
            <TeamSpaceCard />
          </Grid>
        </Grid>
      </MDBox>

      <Footer />
    </DashboardLayout>
  );
}

export default Profile;
