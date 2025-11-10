import React from 'react';
import { Box } from '@mui/material';

interface OvalTableSVGProps {
  width?: number | string;
  height?: number | string;
}

export const OvalTableSVG: React.FC<OvalTableSVGProps> = ({
  width = '100%',
  height = '100%'
}) => {
  return (
    <Box
      sx={{
        width,
        height,
        position: 'relative',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <svg
        width="100%"
        height="100%"
        viewBox="0 0 1200 800"
        preserveAspectRatio="xMidYMid meet"
        style={{ position: 'absolute', top: 0, left: 0 }}
      >
        <defs>
          {/* Felt texture pattern */}
          <pattern id="feltTexture" x="0" y="0" width="4" height="4" patternUnits="userSpaceOnUse">
            <rect width="4" height="4" fill="#0b6b3e" />
            <path d="M 0 0 L 4 4 M 4 0 L 0 4" stroke="#0a5f36" strokeWidth="0.5" opacity="0.1" />
          </pattern>

          {/* Leather texture for rail */}
          <pattern id="leatherTexture" x="0" y="0" width="3" height="3" patternUnits="userSpaceOnUse">
            <rect width="3" height="3" fill="#6b5d52" />
            <circle cx="1.5" cy="1.5" r="0.3" fill="#5a4d44" opacity="0.3" />
          </pattern>

          {/* Gradients for depth */}
          <radialGradient id="feltGradient" cx="50%" cy="50%">
            <stop offset="0%" stopColor="#10854d" />
            <stop offset="50%" stopColor="#0b6b3e" />
            <stop offset="100%" stopColor="#064e3b" />
          </radialGradient>

          <linearGradient id="railGradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="#8b7a6a" />
            <stop offset="30%" stopColor="#6b5d52" />
            <stop offset="70%" stopColor="#5a4d44" />
            <stop offset="100%" stopColor="#4a3d34" />
          </linearGradient>

          {/* Inner shadow for depth */}
          <filter id="innerShadow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur in="SourceAlpha" stdDeviation="8" />
            <feOffset dx="0" dy="4" result="offsetblur" />
            <feComponentTransfer>
              <feFuncA type="linear" slope="0.5" />
            </feComponentTransfer>
            <feMerge>
              <feMergeNode />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>

          {/* Drop shadow for table elevation */}
          <filter id="tableShadow" x="-20%" y="-20%" width="140%" height="140%">
            <feGaussianBlur in="SourceAlpha" stdDeviation="15" />
            <feOffset dx="0" dy="10" result="offsetblur" />
            <feComponentTransfer>
              <feFuncA type="linear" slope="0.4" />
            </feComponentTransfer>
            <feMerge>
              <feMergeNode />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>

          {/* Ambient lighting overlay */}
          <radialGradient id="ambientLight" cx="50%" cy="40%">
            <stop offset="0%" stopColor="#ffffff" stopOpacity="0.15" />
            <stop offset="70%" stopColor="#ffffff" stopOpacity="0.05" />
            <stop offset="100%" stopColor="#000000" stopOpacity="0.1" />
          </radialGradient>
        </defs>

        {/* Table shadow (under the table) */}
        <ellipse
          cx="600"
          cy="410"
          rx="550"
          ry="360"
          fill="rgba(0, 0, 0, 0.4)"
          filter="url(#tableShadow)"
        />

        {/* Outer rail (brown padded border) */}
        <ellipse
          cx="600"
          cy="400"
          rx="560"
          ry="370"
          fill="url(#railGradient)"
          stroke="#3d3228"
          strokeWidth="4"
        />

        {/* Rail leather texture overlay */}
        <ellipse
          cx="600"
          cy="400"
          rx="560"
          ry="370"
          fill="url(#leatherTexture)"
          opacity="0.6"
        />

        {/* Rail padding segments (stitched sections) */}
        {Array.from({ length: 12 }).map((_, i) => {
          const angle = (i / 12) * 2 * Math.PI - Math.PI / 2;
          const rx = 560;
          const ry = 370;
          const x = 600 + rx * Math.cos(angle);
          const y = 400 + ry * Math.sin(angle);

          return (
            <g key={i}>
              {/* Padding button decoration */}
              <circle
                cx={x}
                cy={y}
                r="8"
                fill="#4a3d34"
                stroke="#6b5d52"
                strokeWidth="1.5"
              />
              <circle
                cx={x}
                cy={y}
                r="4"
                fill="#3d3228"
              />
            </g>
          );
        })}

        {/* Inner rail highlight (for 3D effect) */}
        <ellipse
          cx="600"
          cy="395"
          rx="540"
          ry="355"
          fill="none"
          stroke="rgba(139, 122, 106, 0.5)"
          strokeWidth="2"
        />

        {/* Playing surface (green felt) */}
        <ellipse
          cx="600"
          cy="400"
          rx="480"
          ry="310"
          fill="url(#feltGradient)"
          stroke="#064e3b"
          strokeWidth="3"
        />

        {/* Felt texture overlay */}
        <ellipse
          cx="600"
          cy="400"
          rx="480"
          ry="310"
          fill="url(#feltTexture)"
          opacity="0.4"
        />

        {/* Inner shadow on felt */}
        <ellipse
          cx="600"
          cy="400"
          rx="480"
          ry="310"
          fill="rgba(0, 0, 0, 0.2)"
          filter="url(#innerShadow)"
        />

        {/* Ambient lighting overlay */}
        <ellipse
          cx="600"
          cy="400"
          rx="480"
          ry="310"
          fill="url(#ambientLight)"
        />

        {/* Subtle inner border on felt */}
        <ellipse
          cx="600"
          cy="400"
          rx="475"
          ry="305"
          fill="none"
          stroke="rgba(16, 133, 77, 0.4)"
          strokeWidth="1"
        />
      </svg>
    </Box>
  );
};
