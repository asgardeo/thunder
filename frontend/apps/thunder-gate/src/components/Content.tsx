import * as React from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import {
  Cloud,
  ShieldCheck,
  Zap,
  TerminalSquare,
} from "lucide-react";

const items = [
  {
    icon: <Cloud className="text-muted-foreground" />,
    title: "Unified Identity Platform",
    description:
      "Thunder centralizes authentication, authorization, and user management into one secure, scalable identity layer.",
  },
  {
    icon: <ShieldCheck className="text-muted-foreground" />,
    title: "Zero-trust by design",
    description:
      "Leverage built-in OIDC, OAuth 2.0, and adaptive authentication to protect every login and session.",
  },
  {
    icon: <TerminalSquare className="text-muted-foreground" />,
    title: "Lightning-fast developer experience",
    description:
      "Create apps, configure auth flows, and manage tenants in minutes with Thunder CLI, SDKs, and APIs.",
  },
  {
    icon: <Zap className="text-muted-foreground" />,
    title: "Extensible and cloud-ready",
    description:
      "Built with Go and React, Thunder scales effortlessly and integrates with your existing stack and CI/CD pipelines.",
  },
];

export default function Content() {
  return (
    <Stack
      sx={{ flexDirection: 'column', alignSelf: 'center', gap: 4, maxWidth: 450 }}
    >
      <Box sx={{ display: { xs: 'none', md: 'flex' } }}>
        <Typography
          variant="h2">
            ⚡️ Thunder
          </Typography>
      </Box>
      {items.map((item, index) => (
        <Stack key={index} direction="row" sx={{ gap: 2 }}>
          {item.icon}
          <div>
            <Typography gutterBottom sx={{ fontWeight: 'medium' }}>
              {item.title}
            </Typography>
            <Typography variant="body2" sx={{ color: 'text.secondary' }}>
              {item.description}
            </Typography>
          </div>
        </Stack>
      ))}
    </Stack>
  );
}
