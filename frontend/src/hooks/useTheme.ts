import { useState, useEffect } from "react";

type Theme = "dark" | "light";

const STORAGE_KEY = "grow-theme";

function getInitial(): Theme {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "light" || stored === "dark") return stored;
  } catch {}
  return "dark";
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(getInitial);

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
    try {
      localStorage.setItem(STORAGE_KEY, theme);
    } catch {}
  }, [theme]);

  const toggle = () => setThemeState((t) => (t === "dark" ? "light" : "dark"));

  return { theme, toggle } as const;
}
