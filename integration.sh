#!/bin/bash
cd tests/integration/mock_repo
unzip repo.zip # a hack in order to preserve .git

RESULT=$(../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af addition.rb 3 418c8b04c6070b2c08fff937b3cd59f39bb8654c)
[[ $RESULT == "Given code found in addition.rb#L8-8" ]] && echo "[ PASSED ] addition.rb" || echo "[ FAILED ] addition.rb: $RESULT"

RESULT=$(../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af deletion.rb 8 418c8b04c6070b2c08fff937b3cd59f39bb8654c)
[[ $RESULT == "Given code found in deletion.rb#L3-3" ]] && echo "[ PASSED ] deletion.rb" || echo "[ FAILED ] addition.rb: $RESULT"

RESULT=$(../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af deletion.rb 1-3 418c8b04c6070b2c08fff937b3cd59f39bb8654c)
[[ $RESULT == "Given code found in moved.rb#L1-3" ]] && echo "[ PASSED ] moved.rb" || echo "[ FAILED ] addition.rb: $RESULT"