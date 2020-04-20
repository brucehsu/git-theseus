#!/bin/bash
cd tests/integration/mock_repo
unzip -o repo.zip # a hack in order to preserve .git

RESULT=$(../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af addition.rb 3 418c8b04c6070b2c08fff937b3cd59f39bb8654c)
[[ $RESULT == "Given code found in addition.rb#L8-8" ]] && echo "[ PASSED ] Addition" || echo "[ FAILED ] Addition: $RESULT"

RESULT=$(../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af deletion.rb 8 418c8b04c6070b2c08fff937b3cd59f39bb8654c)
[[ $RESULT == "Given code found in deletion.rb#L3-3" ]] && echo "[ PASSED ] Deletion" || echo "[ FAILED ] Deletion: $RESULT"

RESULT=$(../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af deletion.rb 1-3 418c8b04c6070b2c08fff937b3cd59f39bb8654c)
[[ $RESULT == "Given code found in moved.rb#L1-3" ]] && echo "[ PASSED ] Moved" || echo "[ FAILED ] Moved: $RESULT"

RESULT=$(../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af deletion.rb 1-3 a5660be67b5d1b883d44a5f57c7722abe59079af)
[[ $RESULT == "Given code found in deletion.rb#L1-3 [File not changed]" ]] && echo "[ PASSED ] Not changed" || echo "[ FAILED ] Not changed: $RESULT"

RESULT=$(../../../git-theseus a5660be67b5d1b883d44a5f57c7722abe59079af deletion.rb 8 87b47b2af81f2d141107f86a64186c20a39fddc6)
[[ $RESULT == "Given code is not found" ]] && echo "[ PASSED ] Permanent deletion" || echo "[ FAILED ] Permanent deletion: $RESULT"