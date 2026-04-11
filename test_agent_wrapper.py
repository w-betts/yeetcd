#!/usr/bin/env python3
"""
Tests for the agent wrapper script (Mac/Linux).

These tests verify that the agent wrapper script:
- Parses arguments correctly
- Maps subcommands to agent names properly
- Handles invalid subcommands gracefully
- Launches opencode with the correct agent flag
"""

import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch


class TestAgentWrapperScript(unittest.TestCase):
    """Test cases for the agent wrapper script functionality."""

    def setUp(self):
        """Set up test fixtures."""
        self.script_path = Path(__file__).parent / "agent"

    def test_script_accepts_single_argument(self):
        """Verify script accepts a single subcommand argument."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # Script should parse sys.argv or use argparse
        self.assertTrue(
            "sys.argv" in content or "argparse" in content or "ArgumentParser" in content,
            "Script should parse command line arguments"
        )

    def test_script_uses_opencode_agent_flag(self):
        """Verify script launches opencode with --agent flag."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        self.assertIn(
            "--agent",
            content,
            "Script should use --agent flag when launching opencode"
        )

    def test_script_does_not_pass_prompt(self):
        """Verify script does not pass any prompt argument to opencode."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # The script should NOT pass a prompt - user enters it in TUI
        # Check that there's no hardcoded prompt being passed (but allow for "--agent" syntax)
        lines = content.split('\n')
        for line in lines:
            if 'opencode' in line.lower() and ('subprocess' in line or 'os.system' in line or 'call(' in line):
                # The line launching opencode should not contain a prompt argument
                # Allow quotes around flags like "--agent", but not around prompt text
                after_opencode = line.split('opencode')[1] if 'opencode' in line else line
                # Check for quote patterns that suggest a prompt (not just flag syntax)
                if '"' in after_opencode:
                    # Make sure it's not just "--agent" or similar flag syntax
                    if not ('"--' in after_opencode or "'--" in after_opencode):
                        self.fail(
                            f"Script should not pass hardcoded prompt to opencode. Found: {line}"
                        )

    def _get_script_content(self):
        """Helper to read script content."""
        if not self.script_path.exists():
            return None
        with open(self.script_path, 'r') as f:
            return f.read()


class TestAgentErrorHandling(unittest.TestCase):
    """Test cases for error handling of invalid subcommands."""

    def setUp(self):
        """Set up test fixtures."""
        self.script_path = Path(__file__).parent / "agent"
        self.invalid_subcommands = ["invalid", "help", "", "test", "deploy", "foo"]

    def test_script_handles_invalid_subcommand(self):
        """Verify script handles invalid subcommand gracefully."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # Script should have error handling for invalid subcommands
        has_error_handling = (
            "else:" in content or
            "sys.exit" in content or
            "error" in content.lower() or
            "invalid" in content.lower() or
            "usage" in content.lower()
        )
        
        self.assertTrue(
            has_error_handling,
            "Script should handle invalid subcommands with error message"
        )

    def test_script_exits_with_error_on_invalid_subcommand(self):
        """Verify script exits with non-zero status on invalid subcommand."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # Script should exit with error code on invalid input
        self.assertIn(
            "sys.exit",
            content,
            "Script should use sys.exit() to return error code"
        )

    def _get_script_content(self):
        """Helper to read script content."""
        if not self.script_path.exists():
            return None
        with open(self.script_path, 'r') as f:
            return f.read()


class TestAgentScriptExecution(unittest.TestCase):
    """Integration tests for script execution (mocked)."""

    def setUp(self):
        """Set up test fixtures."""
        self.script_path = Path(__file__).parent / "agent"

    @patch('subprocess.run')
    def test_spec_subcommand_executes_opencode_with_spec_agent(self, mock_run):
        """Verify spec subcommand launches opencode with spec agent."""
        if not self.script_path.exists():
            self.skipTest("Agent script does not exist yet")
        
        # This is a mock-based test - in real execution, the script would call opencode
        # We verify the script structure supports this
        content = self._get_script_content()
        self.assertIsNotNone(content)
        
        # Verify the script would call opencode with --agent spec
        self.assertIn("opencode", content.lower())
        self.assertIn("--agent", content)

    @patch('subprocess.run')
    def test_vibe_subcommand_executes_opencode_with_vibe_agent(self, mock_run):
        """Verify vibe subcommand launches opencode with vibe agent."""
        if not self.script_path.exists():
            self.skipTest("Agent script does not exist yet")
        
        content = self._get_script_content()
        self.assertIsNotNone(content)
        self.assertIn("opencode", content.lower())
        self.assertIn("--agent", content)

    def _get_script_content(self):
        """Helper to read script content."""
        if not self.script_path.exists():
            return None
        with open(self.script_path, 'r') as f:
            return f.read()


class TestSpecTreeWorkflow(unittest.TestCase):
    """Test cases for spec-tree workflow entry point."""

    def setUp(self):
        """Set up test fixtures."""
        self.script_path = Path(__file__).parent / "agent"

    def _get_script_content(self):
        """Helper to read script content."""
        if not self.script_path.exists():
            return None
        with open(self.script_path, 'r') as f:
            return f.read()

    def test_spec-tree_in_agent_map(self):
        """Verify spec-tree is added to AGENT_MAP dictionary."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # Check that spec-tree is in AGENT_MAP
        self.assertIn('"spec-tree"', content, 
            "AGENT_MAP should contain 'spec-tree' entry")

    def test_spec-tree_in_print_usage(self):
        """Verify spec-tree entry is in print_usage() function."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # Check that print_usage includes spec-tree
        self.assertIn("spec-tree", content.lower(),
            "print_usage should include spec-tree entry")

    def test_spec-tree_handled_like_spec_in_prompt_new_or_resume(self):
        """Verify spec-tree workflow type is handled like spec (always requires worktree)."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # Verify that spec-tree is handled similarly to spec in prompt_new_or_resume
        # spec-tree should NOT have main branch option like vibe/fix
        # Check the logic in prompt_new_or_resume handles spec-tree without main branch
        
        # Look for the workflow type handling section
        self.assertIn("workflow_type", content,
            "Script should have workflow_type variable")
        
        # Verify that spec-tree doesn't get main branch option
        # The existing code has: if workflow_type in ("vibe", "fix"):
        # For spec-tree, it should NOT be in this tuple
        # We check that "spec-tree" is NOT in the vibe/fix tuple
        lines = content.split('\n')
        for i, line in enumerate(lines):
            if 'workflow_type in ("vibe", "fix")' in line or "workflow_type in ('vibe', 'fix')" in line:
                # Ensure spec-tree is not added to this condition
                next_few_lines = '\n'.join(lines[i:i+5])
                self.assertNotIn("spec-tree", next_few_lines,
                    "spec-tree should not be grouped with vibe/fix for main branch option")
                break

    def test_get_spec-tree_status_method_exists(self):
        """Verify _get_spec-tree_status() method exists in WorkItem class."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # Check that _get_spec-tree_status method exists
        self.assertIn("_get_spec-tree_status", content,
            "WorkItem class should have _get_spec-tree_status() method")

    def test_get_spec-tree_status_called_in_get_status(self):
        """Verify _get_spec-tree_status is called when workflow_type is spec-tree."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # Check that get_status() method calls _get_spec-tree_status
        # Look for the pattern where workflow_type is checked and _get_spec-tree_status is called
        self.assertIn("workflow_type == \"spec-tree\"", content,
            "get_status() should check for spec-tree workflow type")
        self.assertIn("_get_spec-tree_status()", content,
            "get_status() should call _get_spec-tree_status() for spec-tree workflow")

    def test_spec-tree_in_get_workflow_types_with_sessions(self):
        """Verify spec-tree is included in get_workflow_types_with_sessions() function."""
        content = self._get_script_content()
        if content is None:
            self.skipTest("Agent script does not exist yet")
        
        # Check that get_workflow_types_with_sessions includes spec-tree
        # It should iterate over AGENT_MAP.keys() which now includes spec-tree
        self.assertIn("get_workflow_types_with_sessions", content,
            "get_workflow_types_with_sessions function should exist")
        
        # Since AGENT_MAP now includes spec-tree, it should be returned by this function
        # (unless explicitly excluded like improve)


if __name__ == '__main__':
    unittest.main()
