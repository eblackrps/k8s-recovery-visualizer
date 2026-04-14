import type { ReactElement } from "react";

export type NavIconName =
  | "home"
  | "projects"
  | "scan"
  | "live"
  | "complete"
  | "results"
  | "settings";

type IconComponent = () => ReactElement;

const strokeProps = {
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.7,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
};

export const navIcons: Record<NavIconName, IconComponent> = {
  home: () => (
    <svg viewBox="0 0 16 16" aria-hidden="true">
      <path {...strokeProps} d="M2.5 7.2L8 2.8l5.5 4.4" />
      <path {...strokeProps} d="M4.2 6.8V13h7.6V6.8" />
      <path {...strokeProps} d="M6.7 13V9.7h2.6V13" />
    </svg>
  ),
  projects: () => (
    <svg viewBox="0 0 16 16" aria-hidden="true">
      <path {...strokeProps} d="M1.8 4.2h4l1.2 1.6h7.2V12.8H1.8z" />
      <path {...strokeProps} d="M1.8 4.2V3.2h3.6l1 1.2" />
    </svg>
  ),
  scan: () => (
    <svg viewBox="0 0 16 16" aria-hidden="true">
      <circle {...strokeProps} cx="7" cy="7" r="3.8" />
      <path {...strokeProps} d="M9.8 9.8L13.2 13.2" />
      <path {...strokeProps} d="M7 4.8v4.4M4.8 7h4.4" />
    </svg>
  ),
  live: () => (
    <svg viewBox="0 0 16 16" aria-hidden="true">
      <path {...strokeProps} d="M1.5 8h2.3l1.2-2.6 2.1 5.2 1.9-4 1.1 1.4h3.4" />
    </svg>
  ),
  complete: () => (
    <svg viewBox="0 0 16 16" aria-hidden="true">
      <circle {...strokeProps} cx="8" cy="8" r="5.7" />
      <path {...strokeProps} d="M5.3 8.1l1.8 1.8 3.7-4" />
    </svg>
  ),
  results: () => (
    <svg viewBox="0 0 16 16" aria-hidden="true">
      <path {...strokeProps} d="M4 2.5h5.2L12 5.3v8.2H4z" />
      <path {...strokeProps} d="M9.2 2.5v2.8H12" />
      <path {...strokeProps} d="M5.6 8h4.8M5.6 10.2h4.8" />
    </svg>
  ),
  settings: () => (
    <svg viewBox="0 0 16 16" aria-hidden="true">
      <circle {...strokeProps} cx="8" cy="8" r="2.1" />
      <path {...strokeProps} d="M8 1.8v1.4M8 12.8v1.4M3.2 8H1.8M14.2 8h-1.4M12.3 3.7l-1 1M4.7 11.3l-1 1M12.3 12.3l-1-1M4.7 4.7l-1-1" />
    </svg>
  ),
};

export function NavIcon(props: { name: NavIconName }) {
  const Icon = navIcons[props.name];
  return <Icon />;
}
