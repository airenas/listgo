git clone -b r1.7 https://github.com/tensorflow/serving.git
git clone -b r1.7 https://github.com/tensorflow/tensorflow.git

mkdir -p gen
PROTOC_OPTS='-I tensorflow -I serving --go_out=plugins=grpc:gen'

eval "protoc $PROTOC_OPTS serving/tensorflow_serving/apis/*.proto"
eval "protoc $PROTOC_OPTS serving/tensorflow_serving/config/*.proto"
eval "protoc $PROTOC_OPTS serving/tensorflow_serving/util/*.proto"
eval "protoc $PROTOC_OPTS serving/tensorflow_serving/sources/storage_path/*.proto"
eval "protoc $PROTOC_OPTS tensorflow/tensorflow/core/framework/*.proto"
eval "protoc $PROTOC_OPTS tensorflow/tensorflow/core/example/*.proto"
eval "protoc $PROTOC_OPTS tensorflow/tensorflow/core/lib/core/*.proto"
eval "protoc $PROTOC_OPTS tensorflow/tensorflow/core/protobuf/{saver,meta_graph}.proto"

IMP="bitbucket.org\/airenas\/listgo\/internal\/pkg\/tensorflow"
find gen -name '*.go' | xargs sed -i "s/\"tensorflow/\"$IMP/g"

mv gen/tensorflow ../internal/pkg
mv gen/tensorflow_serving ../internal/pkg

rm - r gen
rm - r tensorflow
rm - r serving
