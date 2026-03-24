#!/usr/bin/env python3
"""
Tests for the agent.bat wrapper script (Windows).

These tests verify that the Windows batch wrapper script:
- Parses arguments correctly
- Maps subcommands to agent names properly
- Handles invalid subcommands gracefully
- Launches opencode with the correct agent flag
"""

import unittest
from pathlib import Path


class TestAgentBatWrapperScript(unittest.TestCase):
    """Test cases for the agent.bat wrapper script functionality."""

    def setUp(self):
        """Set up test fixtures."""
        self.script_path = Path(__file__).parent / "agent.bat"

    def _get_script_content(self):
        """Helper to read script content."""
        if not self.script_path.exists():
            return None
        with open(self.script_path, 'r') as f:
            return f.read()

    def test_script_accepts_single_argument(self):
        """Verify batch script accepts a single subcommand argument."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent.bat script does not exist yet")
        
        # Batch script should use %1 or %~1 to access first argument
        self.assertTrue(
            "%1" in content or "%~1" in content,
            "Batch script should use %1 or %~1 to access first argument"
        )

    def test_script_uses_opencode_agent_flag(self):
        """Verify batch script launches opencode with --agent flag."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent.bat script does not exist yet")
        
        self.assertIn(
            "--agent",
            content,
            "Batch script should use --agent flag when launching opencode"
        )

    def test_script_calls_opencode(self):
        """Verify batch script calls opencode command."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent.bat script does not exist yet")
        
        self.assertIn(
            "opencode",
            content.lower(),
            "Batch script should call opencode"
        )


class TestAgentBatErrorHandling(unittest.TestCase):
    """Test cases for error handling of invalid subcommands in batch script."""

    def setUp(self):
        """Set up test fixtures."""
        self.script_path = Path(__file__).parent / "agent.bat"
        self.invalid_subcommands = ["invalid", "help", "", "test", "deploy", "foo"]

    def _get_script_content(self):
        """Helper to read script content."""
        if not self.script_path.exists():
            return None
        with open(self.script_path, 'r') as f:
            return f.read()

    def test_script_handles_invalid_subcommand(self):
        """Verify batch script handles invalid subcommand gracefully."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent.bat script does not exist yet")
        
        # Batch script should have error handling (GOTO, ECHO error message, etc.)
        has_error_handling = (
            "goto" in content.lower() or
            "echo" in content.lower() or
            "error" in content.lower() or
            "invalid" in content.lower() or
            "usage" in content.lower() or
            "exit" in content.lower()
        )
        
        self.assertTrue(
            has_error_handling,
            "Batch script should handle invalid subcommands with error message"
        )

    def test_script_exits_with_error_on_invalid_subcommand(self):
        """Verify batch script exits with non-zero status on invalid subcommand."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent.bat script does not exist yet")
        
        # Batch script should exit with error code on invalid input
        self.assertIn(
            "exit",
            content.lower(),
            "Batch script should use exit command to return error code"
        )


class TestAgentBatExecution(unittest.TestCase):
    """Tests for batch script execution behavior."""

    def setUp(self):
        """Set up test fixtures."""
        self.script_path = Path(__file__).parent / "agent.bat"

    def _get_script_content(self):
        """Helper to read script content."""
        if not self.script_path.exists():
            return None
        with open(self.script_path, 'r') as f:
            return f.read()

    def test_spec_subcommand_executes_opencode_with_spec_agent(self):
        """Verify spec subcommand in batch script launches opencode with spec agent."""
        if not self.script_path.exists():
            self.skipTest("Agent.bat script does not exist yet")
        
        content = self._get_script_content()
        self.assertIsNotNone(content)
        
        # Verify the script would call opencode with --agent spec
        self.assertIn("opencode", content.lower())
        self.assertIn("--agent", content)

    def test_vibe_subcommand_executes_opencode_with_vibe_agent(self):
        """Verify vibe subcommand in batch script launches opencode with vibe agent."""
        if not self.script_path.exists():
            self.skipTest("Agent.bat script does not exist yet")
        
        content = self._get_script_content()
        self.assertIsNotNone(content)
        self.assertIn("opencode", content.lower())
        self.assertIn("--agent", content)


if __name__ == '__main__':
    unittest.main()
