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
      <Paper
        elevation={3}
        sx={{
          width,
          height,
          background: 'linear-gradient(135deg, #1a237e 0%, #283593 100%)',
          border: '2px solid #3949ab',
          borderRadius: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
          '&::before': {
            content: '""',
            position: 'absolute',
            width: '80%',
            height: '80%',
            background: 'repeating-linear-gradient(45deg, transparent, transparent 10px, rgba(255,255,255,0.1) 10px, rgba(255,255,255,0.1) 20px)',
            borderRadius: 1,
          },
        }}
      />
    );
  }

  return (
    <Paper
      elevation={3}
      sx={{
        width,
        height,
        bgcolor: 'white',
        border: '2px solid #ddd',
        borderRadius: 1.5,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'space-between',
        p: 0.5,
        position: 'relative',
      }}
    >
      <Box
        sx={{
          fontSize,
          fontWeight: 'bold',
          color: getSuitColor(parsedCard.suit),
          lineHeight: 1,
        }}
      >
        {parsedCard.rank}
      </Box>
      <Box
        sx={{
          fontSize: fontSize * 1.5,
          color: getSuitColor(parsedCard.suit),
          lineHeight: 1,
        }}
      >
        {parsedCard.suit}
      </Box>
      <Box
        sx={{
          fontSize,
          fontWeight: 'bold',
          color: getSuitColor(parsedCard.suit),
          lineHeight: 1,
        }}
      >
        {parsedCard.rank}
      </Box>
    </Paper>
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
