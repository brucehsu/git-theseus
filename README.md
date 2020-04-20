# git-theseus (POC)
A CLI to check if given section of code is still intact in another commit.

## Usage (Example)
```
> make # You can also use make test-integration to initiate integration test
> cd tests/integration/mock_repo
> unzip -o repo.zip
> ../../../git-theseus
usage: git-theseus [BASE_SHA] [FILE_PATH] [LINE_OR_RANGE] [COMPARE_SHA]
> ../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af addition.rb 3 418c8b04c6070b2c08fff937b3cd59f39bb8654c
Given code found in addition.rb#L8-8
> ../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af deletion.rb 1-3 418c8b04c6070b2c08fff937b3cd59f39bb8654c
Given code found in moved.rb#L1-3
```

## Known Issues/Limitation
- Current implementation does not handle the scenario when the format of the given code, say indentation, has changed
- Content comparison is done line-by-line
- Messy structure and hard-coded values
- The tool does not analyze the semantic, only the lexical form.
