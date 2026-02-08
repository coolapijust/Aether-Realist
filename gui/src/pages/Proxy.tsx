import {
  Box,
  Card,
  CardContent,
  Typography,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Chip,
  Button,
} from '@mui/material';
import { Speed as SpeedIcon } from '@mui/icons-material';
import { useState } from 'react';
import { useCoreStore } from '@/store/coreStore';
import { translations } from '@/lib/i18n';

interface Node {
  id: string;
  name: string;
  address: string;
  latency?: number;
  selected?: boolean;
}

export default function Proxy() {
  const { language } = useCoreStore();
  const t = translations[language];
  const [nodes] = useState<Node[]>([]);

  const [selectedNode, setSelectedNode] = useState('1');

  const getLatencyColor = (latency?: number) => {
    if (!latency) return 'default';
    if (latency < 50) return 'success';
    if (latency < 100) return 'warning';
    return 'error';
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>
          {t.proxy.title}
        </Typography>
        <Button variant="outlined" startIcon={<SpeedIcon />}>
          {t.proxy.btn_speedtest}
        </Button>
      </Box>

      <Card>
        <CardContent sx={{ p: 0 }}>
          <List>
            {nodes.map((node) => (
              <ListItem key={node.id} disablePadding>
                <ListItemButton
                  selected={selectedNode === node.id}
                  onClick={() => setSelectedNode(node.id)}
                >
                  <ListItemText
                    primary={node.name}
                    secondary={node.address}
                    primaryTypographyProps={{ fontWeight: 500 }}
                  />
                  {node.latency && (
                    <Chip
                      label={`${node.latency}ms`}
                      color={getLatencyColor(node.latency) as any}
                      size="small"
                      sx={{ ml: 2 }}
                    />
                  )}
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        </CardContent>
      </Card>
    </Box>
  );
}
