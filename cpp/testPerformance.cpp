#include <stdio.h>
#include <sys/types.h>
#include <sys/time.h>

#include <vector>

#include "tiles.h"

int* nullIntArray = NULL;


void makeValues(int size, float* val) {
	for (int i = 0; i < size; i++) {
		val[i] = 0.8847272 * 10; // TODO this should get random*10
	}
}

double benchmarkHashTiler(int numIterations, int numTilings, int numInputValues) {
	float inputData[numInputValues]; // use the same input data for each iteration below, but each time this function is called the data will be different. (Same as go performance tests.)
	makeValues(numInputValues, inputData);

	timeval t1, t2;
	gettimeofday(&t1, 0);

	int* the_tiles = (int*) malloc(numTilings*sizeof(int));
	for (int i=0; i<numIterations; i++) {
		tiles(the_tiles, numTilings, 16384, inputData, numInputValues, nullIntArray, 0); // TODO compare collision table and number version
	}
	free(the_tiles);

	gettimeofday(&t2, 0);
	double t = t2.tv_sec-t1.tv_sec + ((double)(t2.tv_usec - t1.tv_usec))/1000000.0;
	return t;
}

// The benchmark is run repeatedly with increasing number of iterations (1,2,5,10,20,50,100...)
// until it takes long enough to get a good average.
int numberOfIterationsForAttempt(int attemptNumber) {
	int mostSignificantDigitIndex = attemptNumber%3;
	int powerOfTen = attemptNumber/3;

	int mostSigDigit[] = {1,2,5};
	int num = 1;
	for (; powerOfTen > 0; powerOfTen--) {
		num *= 10;
	}
	return mostSigDigit[mostSignificantDigitIndex] * num;
}

typedef struct {
	double time;
	int numIterations;
} TimeAndNumber;

// benchmarkUntilOneSecond runs with increasing number of iterations until the benchmark takes 1 second to run.
// Returns the time in ns per run.
TimeAndNumber benchmarkUntilOneSecond(int numTilings, int numInputValues) {
	int attemptNumber = 0;
	double time = 0.0;
	int numIterations;
	while (time < 1.0) {
		numIterations = numberOfIterationsForAttempt(attemptNumber);
		time = benchmarkHashTiler(numIterations, numTilings, numInputValues);
		attemptNumber++;
	}
	return TimeAndNumber{time, numIterations};
}

typedef struct {
	const char* const name;
	int values;
	int numTilings;
} BenchmarkSettings;

int main(void) {
	const std::vector<BenchmarkSettings> benchmarks = {
		{"1x1", 1, 1},
		{"4x16", 4, 16},
		{"20x128", 20, 128},
	};

	for (auto benchmark : benchmarks) {
		TimeAndNumber tn = benchmarkUntilOneSecond(benchmark.numTilings, benchmark.values);
		printf("Benchmark %s took %f ns per call to tiles (%d iterations).\n", benchmark.name, tn.time*1000000000/tn.numIterations, tn.numIterations);
	}
	return 0;
}
