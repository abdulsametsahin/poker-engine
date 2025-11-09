import React from 'react';
import { Box, Paper } from '@mui/material';

interface PlayingCardProps {
  card: string;
  size?: 'small' | 'medium' | 'large';
  faceDown?: boolean;
}

const PlayingCard: React.FC<PlayingCardProps> = ({ card, size = 'medium', faceDown = false }) => {
  const getSuitColor = (suit: string) => {
    return suit === '♥' || suit === '♦' ? '#dc143c' : '#000000';
  };

  const parsedCard = parseCard(card);

  const sizeConfig = {
    small: { width: 45, height: 65, fontSize: 16 },
    medium: { width: 60, height: 85, fontSize: 20 },
    large: { width: 75, height: 105, fontSize: 24 },
  };

  const { width, height, fontSize } = sizeConfig[size];

  if (faceDown) {
    return (
      <Box
        sx={{
          width,
          height,
          background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.3) 0%, rgba(79, 70, 229, 0.3) 100%)',
          border: '2px solid rgba(99, 102, 241, 0.5)',
          borderRadius: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
          backdropFilter: 'blur(10px)',
          boxShadow: '0 4px 12px rgba(0, 0, 0, 0.3)',
          '&::before': {
            content: '""',
            position: 'absolute',
            width: '80%',
            height: '80%',
            background: 'repeating-linear-gradient(45deg, transparent, transparent 8px, rgba(255,255,255,0.1) 8px, rgba(255,255,255,0.1) 16px)',
            borderRadius: 1,
          },
        }}
      />
    );
  }

  return (
    <Box
      sx={{
        width,
        height,
        bgcolor: '#ffffff',
        border: '2px solid rgba(0, 0, 0, 0.1)',
        borderRadius: 1.5,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'space-between',
        p: 0.5,
        position: 'relative',
        boxShadow: '0 4px 12px rgba(0, 0, 0, 0.3), 0 2px 4px rgba(0, 0, 0, 0.2)',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 2,
          left: 2,
          right: 2,
          bottom: 2,
          border: '1px solid rgba(0, 0, 0, 0.05)',
          borderRadius: 1,
          pointerEvents: 'none',
        },
      }}
    >
      <Box
        sx={{
          fontSize,
          fontWeight: 900,
          color: getSuitColor(parsedCard.suit),
          lineHeight: 1,
          textShadow: '0 1px 2px rgba(0, 0, 0, 0.1)',
        }}
      >
        {parsedCard.rank}
      </Box>
      <Box
        sx={{
          fontSize: fontSize * 1.5,
          color: getSuitColor(parsedCard.suit),
          lineHeight: 1,
          filter: 'drop-shadow(0 1px 2px rgba(0, 0, 0, 0.1))',
        }}
      >
        {parsedCard.suit}
      </Box>
      <Box
        sx={{
          fontSize,
          fontWeight: 900,
          color: getSuitColor(parsedCard.suit),
          lineHeight: 1,
          textShadow: '0 1px 2px rgba(0, 0, 0, 0.1)',
        }}
      >
        {parsedCard.rank}
      </Box>
    </Box>
  );
};

function parseCard(card: string): { rank: string; suit: string } {
  // Handle formats like "A♠", "As", "A of Spades", "ace_of_spades"
  const suitMap: { [key: string]: string } = {
    '♠': '♠',
    '♥': '♥',
    '♦': '♦',
    '♣': '♣',
    's': '♠',
    'h': '♥',
    'd': '♦',
    'c': '♣',
    'spades': '♠',
    'hearts': '♥',
    'diamonds': '♦',
    'clubs': '♣',
  };

  const rankMap: { [key: string]: string } = {
    '10': '10',
    'T': '10',
    'J': 'J',
    'Q': 'Q',
    'K': 'K',
    'A': 'A',
    'jack': 'J',
    'queen': 'Q',
    'king': 'K',
    'ace': 'A',
  };

  // Try to parse "A♠" or "As" format
  if (card.length === 2 || card.length === 3) {
    const rank = card.slice(0, -1);
    const suitChar = card.slice(-1).toLowerCase();
    return {
      rank: rankMap[rank] || rank,
      suit: suitMap[suitChar] || suitChar,
    };
  }

  // Try to parse "ace_of_spades" format
  if (card.includes('_')) {
    const parts = card.toLowerCase().split('_of_');
    if (parts.length === 2) {
      return {
        rank: rankMap[parts[0]] || parts[0].charAt(0).toUpperCase(),
        suit: suitMap[parts[1]] || '?',
      };
    }
  }

  // Default: return the card as-is
  return { rank: card[0] || '?', suit: card[1] || '?' };
}

export default PlayingCard;
