RUN_NAME="toad_ocr_preprocessor"
if [ -f "./output/bin/${RUN_NAME}" ]; then
    ./output/bin/${RUN_NAME} server
else
  echo "./output/bin/${RUN_NAME} not found! please build first"
fi