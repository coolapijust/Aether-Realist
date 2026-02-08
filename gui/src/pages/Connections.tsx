import {
  Box,
  Card,
  CardContent,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Chip,
} from '@mui/material';
import { Close as CloseIcon } from '@mui/icons-material';
import { useCoreStore } from '@/store/coreStore';
import { translations } from '@/lib/i18n';
import { formatBytes, formatDuration } from '@/utils/format';

export default function Connections() {
  const { streams, closeStream, language } = useCoreStore();
  const t = translations[language];

  const getStateColor = (state: string) => {
    switch (state) {
      case 'active': return 'success';
      case 'opening': return 'warning';
      case 'closing': return 'default';
      default: return 'default';
    }
  };

  const activeStreams = Array.from(streams.values()).filter(s => s.state === 'active' || s.state === 'opening');

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h5" sx={{ mb: 3, fontWeight: 600 }}>
        {t.connections.title}
      </Typography>

      <Card>
        <CardContent sx={{ p: 0 }}>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>{t.connections.col_target}</TableCell>
                  <TableCell>{t.connections.col_status}</TableCell>
                  <TableCell>{t.connections.col_upload}</TableCell>
                  <TableCell>{t.connections.col_download}</TableCell>
                  <TableCell>{t.connections.col_duration}</TableCell>
                  <TableCell>{t.connections.col_action}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {activeStreams.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} align="center" sx={{ py: 4 }}>
                      <Typography color="text.secondary">
                        {t.connections.empty_placeholder}
                      </Typography>
                    </TableCell>
                  </TableRow>
                ) : (
                  activeStreams.map((stream) => (
                    <TableRow key={stream.id}>
                      <TableCell>
                        <Typography variant="body2" fontWeight={500}>
                          {stream.targetHost}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          :{stream.targetPort}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={stream.state === 'active' ? t.connections.status_active : t.connections.status_opening}
                          color={getStateColor(stream.state) as any}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>{formatBytes(stream.bytesSent)}</TableCell>
                      <TableCell>{formatBytes(stream.bytesReceived)}</TableCell>
                      <TableCell>
                        {formatDuration(Date.now() - stream.openedAt)}
                      </TableCell>
                      <TableCell>
                        <IconButton
                          size="small"
                          onClick={() => closeStream(stream.id)}
                        >
                          <CloseIcon fontSize="small" />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>
    </Box>
  );
}
