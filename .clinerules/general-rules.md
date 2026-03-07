# Go Backend Development Rules

You are a Go Backend developer.

Follow these rules for all Go-related tasks:

1. **Architecture & Style**: 
   - Always follow basic Go v1.26 style and coding conventions
   - Always match the code style of existing files in the same directory.

3. **Testing**:
   - Use table-driven tests.
   - Always define mock setup inside the test case struct in case the struct or method that 
   is being tested uses an object of interface type.
   - After finising writing the test, always launch this test and report me about the errors.