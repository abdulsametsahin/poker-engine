import React, { useState, useEffect } from 'react';
import { Box, IconButton, Menu, MenuItem, Typography, Divider } from '@mui/material';
import { GridView, Close } from '@mui/icons-material';
import { useNavigate, useParams } from 'react-router-dom';
import { COLORS, RADIUS } from '../../constants';
import { getActiveTables, removeActiveTable, ActiveTable } from '../../utils/tableManager';

export const TableSwitcher: React.FC = () => {
  const navigate = useNavigate();
  const { tableId: currentTableId } = useParams<{ tableId: string }>();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [activeTables, setActiveTables] = useState<ActiveTable[]>([]);

  useEffect(() => {
    // Load active tables
    setActiveTables(getActiveTables());
    
    // Refresh every 10 seconds
    const interval = setInterval(() => {
      setActiveTables(getActiveTables());
    }, 10000);
    
    return () => clearInterval(interval);
  }, []);

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleSwitchTable = (tableId: string) => {
    navigate(`/game/${tableId}`);
    handleClose();
  };

  const handleRemoveTable = (tableId: string, event: React.MouseEvent) => {
    event.stopPropagation();
    removeActiveTable(tableId);
    setActiveTables(getActiveTables());
  };

  // Don't show if there are no other tables
  const otherTables = activeTables.filter(t => t.tableId !== currentTableId);
  if (otherTables.length === 0) {
    return null;
  }

  return (
    <>
      <IconButton
        onClick={handleClick}
        sx={{
          color: COLORS.primary.main,
          position: 'relative',
          '&:hover': {
            color: COLORS.primary.light,
            background: `${COLORS.primary.main}20`,
          },
        }}
        title="Switch Tables"
      >
        <GridView />
        {otherTables.length > 0 && (
          <Box
            sx={{
              position: 'absolute',
              top: 4,
              right: 4,
              width: 16,
              height: 16,
              borderRadius: '50%',
              background: COLORS.danger.main,
              color: COLORS.text.primary,
              fontSize: '10px',
              fontWeight: 700,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {otherTables.length}
          </Box>
        )}
      </IconButton>

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleClose}
        PaperProps={{
          sx: {
            background: COLORS.background.paper,
            borderRadius: RADIUS.md,
            border: `1px solid ${COLORS.border.main}`,
            minWidth: 250,
            maxHeight: 400,
          },
        }}
      >
        <Box sx={{ px: 2, py: 1, borderBottom: `1px solid ${COLORS.border.main}` }}>
          <Typography variant="body2" sx={{ color: COLORS.text.primary, fontWeight: 600 }}>
            Active Tables
          </Typography>
        </Box>

        {otherTables.map((table) => (
          <MenuItem
            key={table.tableId}
            onClick={() => handleSwitchTable(table.tableId)}
            sx={{
              py: 1.5,
              px: 2,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              '&:hover': {
                background: `${COLORS.primary.main}20`,
              },
            }}
          >
            <Box>
              <Typography
                variant="body2"
                sx={{
                  color: COLORS.text.primary,
                  fontWeight: 500,
                  mb: 0.25,
                }}
              >
                Table {table.tableId.substring(0, 8)}...
              </Typography>
              <Typography
                variant="caption"
                sx={{
                  color: COLORS.text.secondary,
                  fontSize: '11px',
                }}
              >
                {new Date(table.lastActivity).toLocaleTimeString()}
              </Typography>
            </Box>
            <IconButton
              size="small"
              onClick={(e) => handleRemoveTable(table.tableId, e)}
              sx={{
                color: COLORS.text.secondary,
                '&:hover': {
                  color: COLORS.danger.main,
                },
              }}
            >
              <Close fontSize="small" />
            </IconButton>
          </MenuItem>
        ))}

        <Divider sx={{ borderColor: COLORS.border.main }} />

        <MenuItem
          onClick={() => {
            navigate('/lobby');
            handleClose();
          }}
          sx={{
            py: 1.5,
            px: 2,
            '&:hover': {
              background: `${COLORS.secondary.main}20`,
            },
          }}
        >
          <Typography
            variant="body2"
            sx={{
              color: COLORS.secondary.main,
              fontWeight: 600,
            }}
          >
            Join New Table
          </Typography>
        </MenuItem>
      </Menu>
    </>
  );
};

export default TableSwitcher;
