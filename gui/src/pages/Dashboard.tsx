import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Chip,
  Grid,
} from '@mui/material';
import {
  PowerSettingsNew as PowerIcon,
  Speed as SpeedIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material';
import {
  LineChart,
  Line,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Area,
  AreaChart,
} from 'recharts';
import { useCoreStore } from '@/store/coreStore';
import { formatDuration, formatBytes } from '@/utils/format';
import LogPanel from '@/components/LogPanel';

export default function Dashboard() {
  const {
    connectionState,
    coreState,
    currentSession,
    metricsHistory,
    activeStreamCount,
    totalUpload,
    totalDownload,
    systemProxyEnabled,
    toggleSystemProxy,
  } = useCoreStore();

  const isConnected = connectionState === 'connected' && coreState === 'Active';

  // Prepare chart data
  const chartData = metricsHistory.map((m) => ({
    time: new Date(m.timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
    upload: m.upload / 1024 / 1024, // MB
    download: m.download / 1024 / 1024,
    latency: m.latencyMs || 0,
  }));

  const getStatusColor = () => {
    if (coreState === 'Active') return 'success';
    if (coreState === 'Error') return 'error';
    if (coreState === 'Starting' || coreState === 'Rotating') return 'warning';
    return 'default';
  };

  const getStatusText = () => {
    switch (coreState) {
      case 'Active': return '已连接';
      case 'Error': return '错误';
      case 'Starting': return '连接中';
      case 'Rotating': return '轮换中';
      case 'Idle': return '未连接';
      default: return coreState;
    }
  };

  return (
    <Box sx={{ p: 2, height: '100vh', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h6" sx={{ fontWeight: 600 }}>
          首页
        </Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Button
            variant="contained"
            size="small"
            startIcon={<PowerIcon />}
            color={systemProxyEnabled ? 'error' : 'primary'}
            onClick={() => toggleSystemProxy(!systemProxyEnabled)}
          >
            {systemProxyEnabled ? '系统代理已开启' : '开启系统代理'}
          </Button>
          <Button
            variant="outlined"
            size="small"
            startIcon={<SpeedIcon />}
          >
            测速
          </Button>
          <Button
            variant="outlined"
            size="small"
            startIcon={<RefreshIcon />}
          >
            优选
          </Button>
        </Box>
      </Box>

      <Grid container spacing={2} sx={{ flex: 1, minHeight: 0 }}>
        {/* Top Row: Status & Stats */}
        <Grid item xs={12} md={3} sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          {/* Status Card */}
          <Card sx={{ flex: 1 }}>
            <CardContent sx={{ p: '16px !important' }}>
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                连接状态
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                <Box
                  sx={{
                    width: 10,
                    height: 10,
                    borderRadius: '50%',
                    bgcolor: getStatusColor() === 'success' ? 'success.main' :
                      getStatusColor() === 'error' ? 'error.main' :
                        getStatusColor() === 'warning' ? 'warning.main' : 'text.disabled',
                    animation: isConnected ? 'pulse 2s infinite' : 'none',
                  }}
                />
                <Chip
                  label={getStatusText()}
                  color={getStatusColor() as any}
                  size="small"
                  sx={{ height: 24 }}
                />
              </Box>
              <Typography variant="caption" display="block" color="text.secondary">
                {currentSession ? formatDuration(currentSession.uptime) : '-'}
              </Typography>
              <Typography variant="caption" display="block" color="text.secondary">
                {useCoreStore.getState().currentConfig.url || '未配置'}
              </Typography>
            </CardContent>
          </Card>

          {/* Traffic Stats */}
          <Card sx={{ flex: 1 }}>
            <CardContent sx={{ p: '16px !important' }}>
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                实时流量
              </Typography>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end' }}>
                <Box>
                  <Typography variant="body2" color="primary.main">
                    ↑ {formatBytes(totalUpload)}
                  </Typography>
                  <Typography variant="body2" color="success.main">
                    ↓ {formatBytes(totalDownload)}
                  </Typography>
                </Box>
                <Typography variant="caption" color="text.secondary">
                  {activeStreamCount} 连接
                </Typography>
              </Box>
            </CardContent>
          </Card>

          {/* Latency */}
          <Card sx={{ flex: 1 }}>
            <CardContent sx={{ p: '16px !important', height: '100%', display: 'flex', flexDirection: 'column' }}>
              <Typography variant="subtitle2" color="text.secondary">
                延迟
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'baseline', gap: 1, flex: 1 }}>
                <Typography variant="h5" sx={{ fontWeight: 600 }}>
                  {metricsHistory.length > 0
                    ? `${metricsHistory[metricsHistory.length - 1].latencyMs || '-'} ms`
                    : '-'
                  }
                </Typography>
              </Box>
              <Box sx={{ height: 40, mt: 'auto' }}>
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={chartData.slice(-10)}>
                    <Line
                      type="monotone"
                      dataKey="latency"
                      stroke="#3b82f6"
                      strokeWidth={2}
                      dot={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Middle Column: Chart */}
        <Grid item xs={12} md={6} sx={{ display: 'flex' }}>
          <Card sx={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
            <CardContent sx={{ p: '16px !important', flex: 1, display: 'flex', flexDirection: 'column' }}>
              <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                流量趋势
              </Typography>
              <Box sx={{ flex: 1, minHeight: 0 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData}>
                    <XAxis
                      dataKey="time"
                      tick={{ fontSize: 10 }}
                      tickLine={false}
                      interval={2}
                    />
                    <YAxis
                      tick={{ fontSize: 10 }}
                      tickLine={false}
                      tickFormatter={(v) => `${v.toFixed(0)}M`}
                      width={30}
                    />
                    <Area
                      type="monotone"
                      dataKey="download"
                      stackId="1"
                      stroke="#10b981"
                      fill="#10b981"
                      fillOpacity={0.3}
                    />
                    <Area
                      type="monotone"
                      dataKey="upload"
                      stackId="1"
                      stroke="#3b82f6"
                      fill="#3b82f6"
                      fillOpacity={0.3}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Right Column: Logs */}
        <Grid item xs={12} md={3} sx={{ display: 'flex' }}>
          <Box sx={{ flex: 1, height: '100%' }}>
            <LogPanel />
          </Box>
        </Grid>

      </Grid>
    </Box>
  );
}
