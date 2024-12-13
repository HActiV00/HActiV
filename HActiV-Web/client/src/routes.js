// Material Dashboard 2 React layouts
import Dashboard from "layouts/dashboard";
import Notifications from "layouts/notifications";
import EventAlert from "layouts/eventalert";
import Mypage from "layouts/profile";
import AgentControl from "layouts/AgentControl";
import Investigation from "layouts/investigation";

// @mui icons
import Icon from "@mui/material/Icon";

const routes = [
  {
    type: "collapse",
    name: "Dashboard",
    key: "dashboard",
    icon: <Icon fontSize="small">dashboard</Icon>,
    route: "/dashboard",
    component: <Dashboard />,
  },
  {
    type: "collapse",
    name: "Investigation",
    key: "investigation",
    icon: <Icon fontSize="small">investigation</Icon>,
    route: "/dashboard/investigation",
    component: <Investigation/>,
  },
  {
    type: "collapse",
    name: "Notifications",
    key: "notifications",
    icon: <Icon fontSize="small">notifications</Icon>,
    route: "/dashboard/notifications",
    component: <Notifications />,
  },
  {
    type: "collapse",
    name: "Event Alert",
    key: "eventalert",
    icon: <Icon fontSize="small">campaign</Icon>,
    route: "/dashboard/eventalert",
    component: <EventAlert />,
  },
  {
    type: "collapse",
    name: "Mypage",
    key: "profile",
    icon: <Icon fontSize="small">person</Icon>,
    route: "/dashboard/mypage",
    component: <Mypage />,
  },
  {
    type: "collapse",
    name: "Agent Control",
    key: "agent-control",
    icon: <Icon fontSize="small">supervisor_account</Icon>,
    route: "/dashboard/agent-control",
    component: <AgentControl />,
  },
];

export default routes;