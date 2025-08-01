# **App Name**: Kube Sherlock

## Core Features:

- CLI flags: Capture command-line flags for customizing behavior.
- Resource data gathering: Gather data based on user-specified kubernetes resources (pods, deployments, services, etc.)
- Data formatting & Analysis tool: Format gathered resource data for AI consumption, send to an LLM like Gemini, and use its reasoning to determine what additional tooling or context it requires.
- Formatted Output: Output LLM response in a human-readable structured CLI format
- Verbosity Control: Configure a 'verbosity' flag, to decide whether to output detailed execution steps.

## Style Guidelines:

- Primary color: A deep indigo (#4B0082), symbolizing investigation and insight.
- Background color: A very light gray (#F0F0F0), offering a clean backdrop.
- Accent color: A vibrant cyan (#00FFFF) to highlight key findings and suggestions.
- Use 'Inter' sans-serif font for a modern, neutral, readable style, appropriate for displaying CLI outputs and structured information.
- Employ clear and simple icons from a standard icon set (e.g., Font Awesome) to represent Kubernetes resources and troubleshooting steps.
- Structure CLI output into distinct sections, using indentation and separators for clarity.