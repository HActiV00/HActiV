import React, { useState, useEffect, useCallback } from 'react';
import { Snackbar, Alert } from '@mui/material';

interface AlertSystemProps {
  logEventsData: any[];
}

export const AlertSystem: React.FC<AlertSystemProps> = ({ logEventsData }) => {
  const [alertOpen, setAlertOpen] = useState(false);
  const [alertSeverity, setAlertSeverity] = useState<'info' | 'warning' | 'error'>('info');
  const [alertMessage, setAlertMessage] = useState('');

  const checkForAlerts = useCallback((data: any[]) => {
    const varLogEvents = data.filter(event => event.filename.startsWith('/var/log/'));
    if (varLogEvents.length > 0) {
      setAlertSeverity('info');
      setAlertMessage(`${varLogEvents.length} event(s) matching /var/log/* found. Please review these log file accesses.`);
      setAlertOpen(true);
      return;
    }

    const deleteEvents = data.filter(event => 
      event.event_type === 'log_file_delete' && 
      event.process_name === 'rm' && 
      event.filename.startsWith('/var/log/')
    );
    if (deleteEvents.length > 0) {
      setAlertSeverity('warning');
      setAlertMessage(`${deleteEvents.length} log file delete event(s) detected. Please investigate these deletions.`);
      setAlertOpen(true);
      return;
    }

    const highRiskEvents = data.filter(event => 
      event.event_type === 'log_file_delete' && 
      event.process_name === 'rm' && 
      event.filename === '/var/log/auth.log'
    );
    if (highRiskEvents.length > 0) {
      setAlertSeverity('error');
      setAlertMessage(`High risk event detected: Deletion of /var/log/auth.log. Immediate attention required!`);
      setAlertOpen(true);
    }
  }, []);

  useEffect(() => {
    if (logEventsData.length > 0) {
      checkForAlerts(logEventsData);
    }
  }, [logEventsData, checkForAlerts]);

  return (
    <Snackbar
      open={alertOpen}
      autoHideDuration={6000}
      onClose={() => setAlertOpen(false)}
      anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
    >
      <Alert onClose={() => setAlertOpen(false)} severity={alertSeverity} sx={{ width: '100%' }}>
        {alertMessage}
      </Alert>
    </Snackbar>
  );
};