import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "MonitorPro | Proactive IT Operations & Fleet Telemetry",
  description: "Enterprise endpoint monitoring, incident management, advisory AI, policy-controlled remediation, and native low-latency screen sharing.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="light">
      <body className="min-h-screen bg-bg font-sans antialiased text-textMain selection:bg-primary/20 selection:text-primary">
        {children}
      </body>
    </html>
  );
}
