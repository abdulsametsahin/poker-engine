import { useState, useEffect, useMemo } from 'react';

/**
 * Hook to calculate responsive sizes based on container dimensions
 * Maintains consistent aspect ratios across different screen sizes
 */
export const useResponsiveSize = (containerRef: React.RefObject<HTMLElement>) => {
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });

  useEffect(() => {
    const updateDimensions = () => {
      if (containerRef.current) {
        const { width, height } = containerRef.current.getBoundingClientRect();
        setDimensions({ width, height });
      }
    };

    // Initial measurement
    updateDimensions();

    // Update on resize
    const resizeObserver = new ResizeObserver(updateDimensions);
    if (containerRef.current) {
      resizeObserver.observe(containerRef.current);
    }

    return () => {
      resizeObserver.disconnect();
    };
  }, [containerRef]);

  // Calculate responsive sizes based on container dimensions
  const sizes = useMemo(() => {
    const { width, height } = dimensions;
    
    // Base size is determined by the smaller dimension to maintain aspect ratio
    const baseSize = Math.min(width, height);
    
    // Calculate scale factor (1.0 = 1000px base)
    const scale = baseSize / 1000;
    
    return {
      // Container dimensions
      container: { width, height },
      
      // Scale factor for all elements
      scale,
      
      // Table sizing
      table: {
        width: Math.min(width * 0.9, 1200 * scale),
        height: Math.min(height * 0.9, 800 * scale),
      },
      
      // Player seat sizing
      playerSeat: {
        width: 90 * scale,
        height: 60 * scale,
        fontSize: 11 * scale,
        avatarSize: 32 * scale,
      },
      
      // Card sizing
      card: {
        small: {
          width: 32 * scale,
          height: 48 * scale,
          fontSize: 11 * scale,
        },
        medium: {
          width: 50 * scale,
          height: 72 * scale,
          fontSize: 16 * scale,
        },
        large: {
          width: 65 * scale,
          height: 92 * scale,
          fontSize: 20 * scale,
        },
      },
      
      // Pot and betting info
      potDisplay: {
        fontSize: 20 * scale,
        captionSize: 9 * scale,
        padding: 16 * scale,
      },
      
      // Dealer button
      dealerButton: {
        size: 30 * scale,
        fontSize: 12 * scale,
      },
      
      // Action timer
      actionTimer: {
        height: 4 * scale,
        fontSize: 10 * scale,
      },
      
      // Borders and spacing
      borderRadius: {
        sm: 8 * scale,
        md: 12 * scale,
        lg: 16 * scale,
      },
      spacing: {
        xs: 4 * scale,
        sm: 8 * scale,
        md: 16 * scale,
        lg: 24 * scale,
      },
    };
  }, [dimensions]);

  return sizes;
};

/**
 * Utility function to calculate responsive font size
 */
export const getResponsiveFontSize = (baseSize: number, scale: number): string => {
  return `${Math.max(baseSize * scale, baseSize * 0.6)}px`;
};

/**
 * Utility function to maintain aspect ratio
 */
export const getAspectRatioSize = (
  width: number,
  height: number,
  aspectRatio: number
): { width: number; height: number } => {
  const currentRatio = width / height;
  
  if (currentRatio > aspectRatio) {
    // Width is too large, constrain by height
    return {
      width: height * aspectRatio,
      height,
    };
  } else {
    // Height is too large, constrain by width
    return {
      width,
      height: width / aspectRatio,
    };
  }
};
