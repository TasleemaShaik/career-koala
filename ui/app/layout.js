import "./globals.css";
import { Providers } from "./providers";

export const metadata = {
  title: "CareerKoala",
  description: "Career coach UI (Go ADK backend)",
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
