import { useCallback, useEffect, useState } from "react";
import type { MhConFavoriteItem } from "./types";

export const MH_FAVORITES_STORAGE_KEY = "mh_favorites";
export const MH_FAVORITES_CHANGE_EVENT = "favorites-change";

export const dispatchMhFavoritesChange = () => {
  window.dispatchEvent(new CustomEvent(MH_FAVORITES_CHANGE_EVENT));
  localStorage.setItem("_favorites_trigger", Date.now().toString());
  localStorage.removeItem("_favorites_trigger");
};

export function useMhFavorites() {
  const [favorites, setFavorites] = useState<MhConFavoriteItem[]>([]);

  const loadFavorites = useCallback(() => {
    const stored = localStorage.getItem(MH_FAVORITES_STORAGE_KEY);
    if (stored) {
      try {
        setFavorites(JSON.parse(stored));
      } catch (error) {
        console.error("Failed to parse favorites", error);
        setFavorites([]);
      }
      return;
    }

    setFavorites([]);
  }, []);

  useEffect(() => {
    loadFavorites();

    const handleCustomEvent = () => loadFavorites();
    const handleStorageEvent = (event: StorageEvent) => {
      if (event.key === MH_FAVORITES_STORAGE_KEY || event.key === "_favorites_trigger") {
        loadFavorites();
      }
    };

    window.addEventListener(MH_FAVORITES_CHANGE_EVENT, handleCustomEvent);
    window.addEventListener("storage", handleStorageEvent);

    return () => {
      window.removeEventListener(MH_FAVORITES_CHANGE_EVENT, handleCustomEvent);
      window.removeEventListener("storage", handleStorageEvent);
    };
  }, [loadFavorites]);

  const addFavorite = (item: MhConFavoriteItem) => {
    const current = [...favorites];
    if (!current.find(favorite => favorite.id === item.id)) {
      current.push(item);
      localStorage.setItem(MH_FAVORITES_STORAGE_KEY, JSON.stringify(current));
      setFavorites(current);
      dispatchMhFavoritesChange();
    }
  };

  const removeFavorite = (id: string) => {
    const current = favorites.filter(favorite => favorite.id !== id);
    localStorage.setItem(MH_FAVORITES_STORAGE_KEY, JSON.stringify(current));
    setFavorites(current);
    dispatchMhFavoritesChange();
  };

  return {
    favorites,
    addFavorite,
    removeFavorite
  };
}
