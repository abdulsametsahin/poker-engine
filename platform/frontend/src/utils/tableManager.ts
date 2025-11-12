/**
 * Utility functions for managing active table sessions
 * Allows users to track and switch between multiple active poker tables
 */

const ACTIVE_TABLES_KEY = 'poker_active_tables';

export interface ActiveTable {
  tableId: string;
  joinedAt: number;
  lastActivity: number;
}

/**
 * Get all active tables for the current user
 */
export const getActiveTables = (): ActiveTable[] => {
  try {
    const stored = localStorage.getItem(ACTIVE_TABLES_KEY);
    if (!stored) return [];
    
    const tables: ActiveTable[] = JSON.parse(stored);
    // Filter out tables older than 24 hours
    const cutoff = Date.now() - (24 * 60 * 60 * 1000);
    return tables.filter(t => t.lastActivity > cutoff);
  } catch (error) {
    console.error('Error loading active tables:', error);
    return [];
  }
};

/**
 * Add or update a table in active tables list
 */
export const addActiveTable = (tableId: string): void => {
  try {
    const tables = getActiveTables();
    const existing = tables.find(t => t.tableId === tableId);
    
    if (existing) {
      existing.lastActivity = Date.now();
    } else {
      tables.push({
        tableId,
        joinedAt: Date.now(),
        lastActivity: Date.now(),
      });
    }
    
    // Keep only the 10 most recent tables
    const sorted = tables.sort((a, b) => b.lastActivity - a.lastActivity).slice(0, 10);
    localStorage.setItem(ACTIVE_TABLES_KEY, JSON.stringify(sorted));
  } catch (error) {
    console.error('Error adding active table:', error);
  }
};

/**
 * Remove a table from active tables list
 */
export const removeActiveTable = (tableId: string): void => {
  try {
    const tables = getActiveTables().filter(t => t.tableId !== tableId);
    localStorage.setItem(ACTIVE_TABLES_KEY, JSON.stringify(tables));
  } catch (error) {
    console.error('Error removing active table:', error);
  }
};

/**
 * Update last activity timestamp for a table
 */
export const updateTableActivity = (tableId: string): void => {
  try {
    const tables = getActiveTables();
    const table = tables.find(t => t.tableId === tableId);
    if (table) {
      table.lastActivity = Date.now();
      localStorage.setItem(ACTIVE_TABLES_KEY, JSON.stringify(tables));
    }
  } catch (error) {
    console.error('Error updating table activity:', error);
  }
};
