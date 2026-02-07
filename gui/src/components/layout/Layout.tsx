import { useEffect } from 'react';
import { 
  Box, 
  Drawer, 
  List, 
  ListItem, 
  ListItemButton,
  ListItemIcon, 
  ListItemText,
  Toolbar,
  Typography,
  IconButton,
  Divider,
  Chip,
} from '@mui/material';
import {
  Home as HomeIcon,
  Language as ProxyIcon,
  Shield as RulesIcon,
  SwapHoriz as ConnectionsIcon,
  Article as LogsIcon,
  Settings as SettingsIcon,
  Brightness4 as DarkModeIcon,
  Brightness7 as LightModeIcon,
  Power as PowerIcon,
} from '@mui/icons-material';
import { useTheme } from '@mui/material/styles';
import { useCoreStore } from '@/store/coreStore';

const drawerWidth = 240;

interface LayoutProps {
  children: React.ReactNode;
  currentPage: string;
  onPageChange: (page: string) => void;
  darkMode: boolean;
  onToggleDarkMode: () => void;
}

const menuItems = [
  { id: 'dashboard', label: '首页', icon: HomeIcon },
  { id: 'proxy', label: '代理', icon: ProxyIcon },
  { id: 'rules', label: '规则', icon: RulesIcon },
  { id: 'connections', label: '连接', icon: ConnectionsIcon },
  { id: 'logs', label: '日志', icon: LogsIcon },
  { id: 'settings', label: '设置', icon: SettingsIcon },
];

export default function Layout({ 
  children, 
  currentPage, 
  onPageChange,
  darkMode,
  onToggleDarkMode,
}: LayoutProps) {
  const theme = useTheme();
  const { connectionState, connect, disconnect } = useCoreStore();
  
  useEffect(() => {
    // Auto-connect on mount
    connect();
    return () => disconnect();
  }, []);

  const getConnectionColor = () => {
    if (connectionState === 'connected') return 'success';
    if (connectionState === 'connecting') return 'warning';
    return 'error';
  };

  return (
    <Box sx={{ display: 'flex', height: '100vh' }}>
      {/* Sidebar */}
      <Drawer
        variant="permanent"
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: drawerWidth,
            boxSizing: 'border-box',
            borderRight: `1px solid ${theme.palette.divider}`,
          },
        }}
      >
        <Toolbar sx={{ px: 2, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              Aether
            </Typography>
            <Typography variant="caption" sx={{ ml: 0.5, opacity: 0.7 }}>
              Realist
            </Typography>
          </Box>
          <Chip 
            size="small" 
            color={getConnectionColor() as any}
            sx={{ 
              width: 8, 
              height: 8, 
              minWidth: 8,
              '& .MuiChip-label': { display: 'none' }
            }}
          />
        </Toolbar>
        
        <Divider />
        
        <List sx={{ px: 1, py: 1 }}>
          {menuItems.map((item) => {
            const Icon = item.icon;
            return (
              <ListItem key={item.id} disablePadding>
                <ListItemButton
                  selected={currentPage === item.id}
                  onClick={() => onPageChange(item.id)}
                >
                  <ListItemIcon>
                    <Icon />
                  </ListItemIcon>
                  <ListItemText primary={item.label} />
                </ListItemButton>
              </ListItem>
            );
          })}
        </List>
        
        <Box sx={{ flexGrow: 1 }} />
        
        <Divider />
        
        <Box sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <IconButton onClick={onToggleDarkMode} color="inherit" size="small">
            {darkMode ? <LightModeIcon /> : <DarkModeIcon />}
          </IconButton>
          
          {connectionState === 'connected' ? (
            <IconButton onClick={disconnect} color="error" size="small">
              <PowerIcon />
            </IconButton>
          ) : (
            <IconButton onClick={connect} color="success" size="small">
              <PowerIcon />
            </IconButton>
          )}
        </Box>
      </Drawer>

      {/* Main content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          bgcolor: 'background.default',
          overflow: 'auto',
        }}
      >
        {children}
      </Box>
    </Box>
  );
}
